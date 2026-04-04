const DFU_DETACH = 0x00;
const DFU_DOWNLOAD = 0x01;
const DFU_GETSTATUS = 0x03;
const DFU_CLRSTATUS = 0x04;
const DFU_GETSTATE = 0x05;
const DFU_ABORT = 0x06;

const STATE_DFU_IDLE = 2;
const STATE_DFU_DOWNLOAD_IDLE = 5;
const STATE_DFU_MANIFEST = 7;
const STATE_DFU_DN_BUSY = 4;
const STATE_DFU_ERROR = 10;

const STATUS_OK = 0x00;

const DFUSE_SET_ADDRESS = 0x21;
const DFUSE_ERASE_SECTOR = 0x41;

const DFUSE_PROTOCOL = 0x02;

const DT_INTERFACE = 4;
const DT_DFU_FUNCTIONAL = 0x21;
const DT_CONFIGURATION = 0x02;
const DT_STRING = 0x03;
const USB_CLASS_APP_SPECIFIC = 0xfe;
const USB_SUBCLASS_DFU = 0x01;
const GET_DESCRIPTOR = 0x06;

const DFUSE_DEVICE_FILTER = {
  classCode: USB_CLASS_APP_SPECIFIC,
  subclassCode: USB_SUBCLASS_DFU,
  protocolCode: DFUSE_PROTOCOL,
};

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

const SECTOR_MULTIPLIERS = {
  " ": 1,
  B: 1,
  K: 1024,
  M: 1024 * 1024,
};

export function parseMemoryDescriptor(desc) {
  const nameEndIndex = desc.indexOf("/");
  if (!desc.startsWith("@") || nameEndIndex === -1)
    throw new Error(`Not a DfuSe memory descriptor: "${desc}"`);

  const name = desc.substring(1, nameEndIndex).trim();
  const segmentString = desc.substring(nameEndIndex);

  const segments = [];

  const contiguousRe = /\/\s*(0x[0-9a-fA-F]{1,8})\s*\/(\s*[0-9]+\s*\*\s*[0-9]+\s?[ BKM]\s*[abcdefg]\s*,?\s*)+/g;
  for (let cm; (cm = contiguousRe.exec(segmentString)); ) {
    const sectorRe = /([0-9]+)\s*\*\s*([0-9]+)\s?([ BKM])\s*([abcdefg])\s*,?\s*/g;
    let startAddress = parseInt(cm[1], 16);

    for (let sm; (sm = sectorRe.exec(cm[0])); ) {
      const sectorCount = parseInt(sm[1], 10);
      const sectorSize = parseInt(sm[2], 10) * (SECTOR_MULTIPLIERS[sm[3]] ?? 0);
      const properties = sm[4].charCodeAt(0) - "a".charCodeAt(0) + 1;

      segments.push({
        start: startAddress,
        sectorSize,
        end: startAddress + sectorSize * sectorCount,
        readable: (properties & (1 << 0)) !== 0,
        erasable: (properties & (1 << 1)) !== 0,
        writable: (properties & (1 << 2)) !== 0,
      });

      startAddress += sectorSize * sectorCount;
    }
  }

  return {
    name,
    segments,
  };
}

function parseFunctionalDescriptor(data) {
  return {
    bLength: data.getUint8(0),
    bDescriptorType: data.getUint8(1),
    bmAttributes: data.getUint8(2),
    wDetachTimeOut: data.getUint16(3, true),
    wTransferSize: data.getUint16(5, true),
    bcdDFUVersion: data.getUint16(7, true),
  };
}

function parseInterfaceDescriptor(data) {
  return {
    bLength: data.getUint8(0),
    bDescriptorType: data.getUint8(1),
    bInterfaceNumber: data.getUint8(2),
    bAlternateSetting: data.getUint8(3),
    bNumEndpoints: data.getUint8(4),
    bInterfaceClass: data.getUint8(5),
    bInterfaceSubClass: data.getUint8(6),
    bInterfaceProtocol: data.getUint8(7),
    iInterface: data.getUint8(8),
    descriptors: [],
  };
}

function parseSubDescriptors(descriptorData) {
  const descriptors = [];
  let remaining = descriptorData;
  let currIntf = null;
  let inDfuIntf = false;

  while (remaining.byteLength > 2) {
    const bLength = remaining.getUint8(0);
    const bDescriptorType = remaining.getUint8(1);
    const descData = new DataView(remaining.buffer.slice(remaining.byteOffset, remaining.byteOffset + bLength));

    if (bDescriptorType === DT_INTERFACE) {
      currIntf = parseInterfaceDescriptor(descData);
      inDfuIntf = currIntf.bInterfaceClass === USB_CLASS_APP_SPECIFIC && currIntf.bInterfaceSubClass === USB_SUBCLASS_DFU;
      descriptors.push(currIntf);
    }
    else if (inDfuIntf && bDescriptorType === DT_DFU_FUNCTIONAL) {
      const funcDesc = parseFunctionalDescriptor(descData);
      descriptors.push(funcDesc);
      currIntf?.descriptors.push(funcDesc);
    }
    else {
      const desc = { bLength, bDescriptorType, descData };
      descriptors.push(desc);
      currIntf?.descriptors.push(desc);
    }

    remaining = new DataView(remaining.buffer, remaining.byteOffset + bLength);
  }

  return descriptors;
}

function parseConfigurationDescriptor(data) {
  const descriptorData = new DataView(data.buffer, data.byteOffset + 9);
  return {
    bConfigurationValue: data.getUint8(5),
    descriptors: parseSubDescriptors(descriptorData),
  };
}

export class DfuSe {
  #device;
  #interfaces;
  #properties;
  #currentInterface;
  #memoryInfo;
  #onDeviceDisconnect;
  #disconnected = false;
  #onDisconnect;

  constructor() {
    if (typeof navigator === "undefined" || !navigator.usb)
      throw new Error(
        "WebUSB is not supported in this browser. Please try a Chromium-based browser " +
        "like Google Chrome, Microsoft Edge, etc."
      );
  }

  get device() {
    return this.#device;
  }

  get memoryInfo() {
    return this.#memoryInfo;
  }

  get interfaces() {
    return this.#interfaces;
  }

  get properties() {
    return this.#properties;
  }

  async connect(onDisconnect, vendorId, productId) {
    const filter = { ...DFUSE_DEVICE_FILTER };
    if (vendorId !== undefined)
      filter.vendorId = vendorId;
    if (productId !== undefined)
      filter.productId = productId;

    this.#device = await navigator.usb.requestDevice({ filters: [filter] });
    this.#interfaces = await this.#findDfuInterfaces();

    if (!this.#device.opened)
      await this.#device.open();

    const desc = await this.#getDFUDescriptorProperties();
    if (desc)
      this.#properties = desc;

    this.#onDeviceDisconnect = onDisconnect;
    this.#disconnected = false;
    this.#onDisconnect = (event) => {
      if (event.device === this.#device) {
        this.#disconnected = true;
        this.#onDeviceDisconnect();
      }
    };
    navigator.usb.addEventListener("disconnect", this.#onDisconnect);

    return this.#interfaces.map((i) => {
      let name = null;
      let totalSize = 0;
      if (i.name) {
        const mem = parseMemoryDescriptor(i.name);
        name = mem.name;
        for (const seg of mem.segments)
          totalSize += seg.end - seg.start;
      }
      return {
        alternateSetting: i.alternate.alternateSetting,
        name,
        totalSize,
      };
    });
  }

  async configure(alternateSetting = 0) {
    if (!this.#device?.opened)
      throw new Error("Device not connected");

    const intrf = this.#interfaces.find((i) => i.alternate.alternateSetting === alternateSetting);
    if (!intrf)
      throw new Error(`Alternate setting ${alternateSetting} not found`);

    this.#currentInterface = intrf;
    this.#memoryInfo = this.#currentInterface.name
      ? parseMemoryDescriptor(this.#currentInterface.name)
      : null;

    await this.#openInterface();
  }

  async close() {
    this.#cleanupDisconnectListener();
    if (this.#device?.opened)
      await this.#device.close();
  }

  async write(data, startAddress, onProgress, onLog) {
    if (!(data instanceof Uint8Array))
      throw new Error("data must be a Uint8Array");

    if (typeof startAddress !== "number" || !Number.isInteger(startAddress) || startAddress < 0)
      throw new Error("startAddress must be a non-negative integer");

    if (!this.#memoryInfo?.segments)
      throw new Error("No memory map available");

    if (!this.#properties?.CanDownload)
      throw new Error("Device does not support download (write) operations");

    const xferSize = this.#properties.TransferSize;
    if (!xferSize)
      throw new Error("Transfer size not available from device");

    if (data.byteLength !== 0) {
      const endAddress = startAddress + data.byteLength;
      let addr = startAddress;
      while (addr < endAddress) {
        const seg = this.#getSegment(addr);
        if (!seg)
          throw new Error(`Address 0x${addr.toString(16)} outside of memory map bounds`);
        if (!seg.writable)
          throw new Error(`Segment at 0x${seg.start.toString(16)} is not writable`);
        addr = seg.end;
      }
    }

    const expectedSize = data.byteLength;

    onLog(`Writing ${expectedSize} bytes to 0x${startAddress.toString(16)} (transfer size: ${xferSize})`);

    await this.#abortToIdle();

    onLog("Erasing...");
    onProgress("erase", 0, expectedSize);
    await this.#erase(startAddress, expectedSize, (sent, total) => {
      onProgress("erase", sent, total);
    });
    onLog("Erase complete");

    onLog("Writing...");
    let bytesSent = 0;
    let address = startAddress;

    onProgress("write", 0, expectedSize);
    while (bytesSent < expectedSize) {
      const chunkSize = Math.min(expectedSize - bytesSent, xferSize);
      const chunk = data.subarray(bytesSent, bytesSent + chunkSize);

      await this.#dfuseCommand(DFUSE_SET_ADDRESS, address, 4);
      const bytesWritten = await this.#download(chunk, 2);
      const status = await this.#pollUntilIdle(STATE_DFU_DOWNLOAD_IDLE);

      if (status.status !== STATUS_OK)
        throw new Error(`DFU DOWNLOAD failed state=${status.state}, status=${status.status}`);

      address += chunkSize;
      bytesSent += bytesWritten;
      onProgress("write", bytesSent, expectedSize);
    }
    onLog(`Write complete (${bytesSent} bytes)`);

    onLog("Manifesting...");
    await this.#dfuseCommand(DFUSE_SET_ADDRESS, startAddress, 4);
    await this.#download(new ArrayBuffer(0), 0);
    await this.#pollUntil((state) => state === STATE_DFU_MANIFEST);

    onLog("Waiting for device to disconnect...");
    await this.#waitDisconnected(5000);
    await this.close();
    onLog("Done");
  }

  async detach() {
    await this.#requestOut(DFU_DETACH, undefined, 1000);
  }

  async abort() {
    await this.#requestOut(DFU_ABORT);
  }

  async getState() {
    const data = await this.#requestIn(DFU_GETSTATE, 1);
    return data.getUint8(0);
  }

  async getStatus() {
    const data = await this.#requestIn(DFU_GETSTATUS, 6);
    return {
      status: data.getUint8(0),
      pollTimeout: data.getUint32(1, true) & 0xffffff,
      state: data.getUint8(4),
    };
  }

  async clearStatus() {
    await this.#requestOut(DFU_CLRSTATUS);
  }

  #waitDisconnected(timeout) {
    if (this.#disconnected) {
      this.#cleanupDisconnectListener();
      return Promise.resolve();
    }

    const device = this.#device;

    return new Promise((resolve, reject) => {
      let timeoutID;

      const onDisconnect = (event) => {
        if (event.device !== device)
          return;

        if (timeoutID)
          clearTimeout(timeoutID);
        navigator.usb.removeEventListener("disconnect", onDisconnect);
        this.#cleanupDisconnectListener();
        resolve();
      };

      navigator.usb.addEventListener("disconnect", onDisconnect);

      if (timeout > 0) {
        timeoutID = setTimeout(() => {
          navigator.usb.removeEventListener("disconnect", onDisconnect);
          reject(new Error("Device did not disconnect after manifestation"));
        }, timeout);
      }
    });
  }

  #cleanupDisconnectListener() {
    if (this.#onDisconnect) {
      navigator.usb.removeEventListener("disconnect", this.#onDisconnect);
      this.#onDisconnect = null;
    }
  }

  get #intfNumber() {
    if (!this.#currentInterface)
      throw new Error("No interface selected");

    return this.#currentInterface.interface.interfaceNumber;
  }

  #assertConnected() {
    if (this.#disconnected)
      throw new Error("Device disconnected unexpectedly");
  }

  async #requestOut(bRequest, data, wValue = 0) {
    this.#assertConnected();
    const result = await this.#device.controlTransferOut(
      {
        requestType: "class",
        recipient: "interface",
        request: bRequest,
        value: wValue,
        index: this.#intfNumber,
      },
      data
    );

    if (result.status !== "ok")
      throw new Error(`ControlTransferOut failed: ${result.status}`);

    return result.bytesWritten;
  }

  async #requestIn(bRequest, wLength, wValue = 0) {
    this.#assertConnected();
    const result = await this.#device.controlTransferIn(
      {
        requestType: "class",
        recipient: "interface",
        request: bRequest,
        value: wValue,
        index: this.#intfNumber,
      },
      wLength
    );

    if (result.status !== "ok" || !result.data)
      throw new Error(`ControlTransferIn failed: ${result.status}`);

    return result.data;
  }

  #download(data, blockNum) {
    return this.#requestOut(DFU_DOWNLOAD, data, blockNum);
  }

  async #abortToIdle() {
    await this.abort();
    let state = await this.getState();
    if (state === STATE_DFU_ERROR) {
      await this.clearStatus();
      state = await this.getState();
    }
    if (state !== STATE_DFU_IDLE)
      throw new Error(`Failed to return to idle state after abort: state ${state}`);
  }

  async #pollUntil(predicate) {
    let dfuStatus = await this.getStatus();
    while (!predicate(dfuStatus.state) && dfuStatus.state !== STATE_DFU_ERROR) {
      await sleep(dfuStatus.pollTimeout);
      dfuStatus = await this.getStatus();
    }
    return dfuStatus;
  }

  #pollUntilIdle(idleState) {
    return this.#pollUntil((state) => state === idleState);
  }

  #getSegment(addr) {
    if (!this.#memoryInfo?.segments)
      throw new Error("No memory map information available");

    for (const segment of this.#memoryInfo.segments) {
      if (segment.start <= addr && addr < segment.end)
        return segment;
    }
    return null;
  }

  #getSectorStart(addr, segment) {
    segment ??= this.#getSegment(addr);
    if (!segment)
      throw new Error(`Address 0x${addr.toString(16)} outside of memory map`);

    const sectorIndex = Math.floor((addr - segment.start) / segment.sectorSize);
    return segment.start + sectorIndex * segment.sectorSize;
  }

  #getSectorEnd(addr, segment) {
    segment ??= this.#getSegment(addr);
    if (!segment)
      throw new Error(`Address 0x${addr.toString(16)} outside of memory map`);

    const sectorIndex = Math.floor((addr - segment.start) / segment.sectorSize);
    return segment.start + (sectorIndex + 1) * segment.sectorSize;
  }

  async #erase(startAddr, length, onProgress) {
    let segment = this.#getSegment(startAddr);
    let addr = this.#getSectorStart(startAddr, segment);
    const endAddr = this.#getSectorEnd(startAddr + length - 1);
    let bytesErased = 0;
    const bytesToErase = endAddr - addr;

    while (addr < endAddr) {
      if ((segment?.end ?? 0) <= addr)
        segment = this.#getSegment(addr);

      if (!segment?.erasable) {
        bytesErased = Math.min(bytesErased + (segment?.end ?? 0) - addr, bytesToErase);
        addr = segment?.end ?? 0;
      }
      else {
        const sectorIndex = Math.floor((addr - segment.start) / segment.sectorSize);
        const sectorAddr = segment.start + sectorIndex * segment.sectorSize;
        await this.#dfuseCommand(DFUSE_ERASE_SECTOR, sectorAddr, 4);
        addr = sectorAddr + segment.sectorSize;
        bytesErased += segment.sectorSize;
      }

      onProgress(bytesErased, bytesToErase);
    }
  }

  async #dfuseCommand(command, param = 0x00, len = 1) {
    const payload = new ArrayBuffer(len + 1);
    const view = new DataView(payload);
    view.setUint8(0, command);

    if (len === 1)
      view.setUint8(1, param);
    else if (len === 4)
      view.setUint32(1, param, true);
    else
      throw new Error(`Unsupported DfuSe command length: ${len}`);

    await this.#download(payload, 0);

    const status = await this.#pollUntil((state) => state !== STATE_DFU_DN_BUSY);
    if (status.status !== STATUS_OK)
      throw new Error("Special DfuSe command failed");
  }

  async #getDFUDescriptorProperties() {
    const data = await this.#readConfigurationDescriptor(0);
    const configDesc = parseConfigurationDescriptor(data);
    const configValue = this.#device.configuration?.configurationValue;

    let funcDesc = null;
    if (configDesc.bConfigurationValue === configValue) {
      for (const desc of configDesc.descriptors) {
        if (desc.bDescriptorType === DT_DFU_FUNCTIONAL && "bcdDFUVersion" in desc) {
          funcDesc = desc;
          break;
        }
      }
    }

    if (!funcDesc)
      return null;

    return {
      WillDetach: (funcDesc.bmAttributes & 0x08) !== 0,
      ManifestationTolerant: (funcDesc.bmAttributes & 0x04) !== 0,
      CanUpload: (funcDesc.bmAttributes & 0x02) !== 0,
      CanDownload: (funcDesc.bmAttributes & 0x01) !== 0,
      TransferSize: funcDesc.wTransferSize,
      DetachTimeOut: funcDesc.wDetachTimeOut,
      DFUVersion: funcDesc.bcdDFUVersion,
    };
  }

  async #findDfuInterfaces() {
    const interfaces = [];

    for (const conf of this.#device.configurations) {
      for (const intf of conf.interfaces) {
        for (const alt of intf.alternates) {
          if (
            alt.interfaceClass === USB_CLASS_APP_SPECIFIC &&
            alt.interfaceSubclass === USB_SUBCLASS_DFU &&
            alt.interfaceProtocol === DFUSE_PROTOCOL
          ) {
            interfaces.push({
              configuration: conf,
              interface: intf,
              alternate: alt,
              name: alt.interfaceName,
            });
          }
        }
      }
    }

    await this.#fixInterfaceNames(interfaces);

    return interfaces;
  }

  async #fixInterfaceNames(interfaces) {
    if (!interfaces.some((intf) => intf.name == null))
      return;

    await this.#device.open();
    await this.#device.selectConfiguration(1);

    const mapping = await this.#readInterfaceNames();

    for (const intf of interfaces) {
      if (intf.name !== null)
        continue;

      const configIndex = intf.configuration.configurationValue;
      const intfNumber = intf.interface.interfaceNumber;
      const alt = intf.alternate.alternateSetting;
      intf.name = mapping?.[configIndex]?.[intfNumber]?.[alt]?.toString();
    }
  }

  async #readStringDescriptor(index, langID = 0) {
    const wValue = (DT_STRING << 8) | index;

    const setup = {
      requestType: "standard",
      recipient: "device",
      request: GET_DESCRIPTOR,
      value: wValue,
      index: langID,
    };

    let result = await this.#device.controlTransferIn(setup, 1);

    if (result.data && result.status === "ok") {
      const bLength = result.data.getUint8(0);
      result = await this.#device.controlTransferIn(setup, bLength);

      if (result.data && result.status === "ok") {
        const len = (bLength - 2) / 2;
        const u16Words = [];
        for (let i = 0; i < len; i++)
          u16Words.push(result.data.getUint16(2 + i * 2, true));
        if (langID === 0)
          return u16Words;
        return String.fromCharCode(...u16Words);
      }
    }

    throw new Error(`Failed to read string descriptor ${index}: ${result.status}`);
  }

  async #readInterfaceNames() {
    const configs = {};
    const allStringIndices = new Set();

    for (let configIndex = 0; configIndex < this.#device.configurations.length; configIndex++) {
      const rawConfig = await this.#readConfigurationDescriptor(configIndex);
      const configDesc = parseConfigurationDescriptor(rawConfig);
      const configValue = configDesc.bConfigurationValue;
      configs[configValue] = {};

      for (const desc of configDesc.descriptors) {
        if (desc.bDescriptorType !== DT_INTERFACE)
          continue;

        configs[configValue][desc.bInterfaceNumber] ??= {};
        configs[configValue][desc.bInterfaceNumber][desc.bAlternateSetting] = desc.iInterface;
        if (desc.iInterface > 0)
          allStringIndices.add(desc.iInterface);
      }
    }

    const strings = {};
    for (const index of allStringIndices) {
      try {
        strings[index] = await this.#readStringDescriptor(index, 0x0409);
      }
      catch {
        strings[index] = null;
      }
    }

    for (const config of Object.values(configs))
      for (const intf of Object.values(config))
        for (const alt in intf)
          intf[alt] = strings[intf[alt]];

    return configs;
  }

  async #readConfigurationDescriptor(index) {
    const wValue = (DT_CONFIGURATION << 8) | index;

    const setup = {
      requestType: "standard",
      recipient: "device",
      request: GET_DESCRIPTOR,
      value: wValue,
      index: 0,
    };

    const sizeResult = await this.#device.controlTransferIn(setup, 4);
    if (!sizeResult.data || sizeResult.status !== "ok")
      throw new Error(`controlTransferIn error: ${sizeResult.status}`);

    const wLength = sizeResult.data.getUint16(2, true);
    const descriptor = await this.#device.controlTransferIn(setup, wLength);
    if (!descriptor.data || descriptor.status !== "ok")
      throw new Error(`controlTransferIn error: ${descriptor.status}`);

    return descriptor.data;
  }

  async #openInterface() {
    if (!this.#currentInterface)
      throw new Error("No interface selected");

    const confValue = this.#currentInterface.configuration.configurationValue;

    if (!this.#device.configuration || this.#device.configuration.configurationValue !== confValue)
      await this.#device.selectConfiguration(confValue);

    if (!this.#device.configuration)
      throw new Error(`Couldn't select configuration '${confValue}'`);

    const intfNumber = this.#currentInterface.interface.interfaceNumber;
    if (!this.#device.configuration.interfaces[intfNumber]?.claimed)
      await this.#device.claimInterface(intfNumber);

    const altSetting = this.#currentInterface.alternate.alternateSetting;
    const intf = this.#device.configuration.interfaces[intfNumber];
    if (!intf?.alternate || intf.alternate.alternateSetting !== altSetting)
      await this.#device.selectAlternateInterface(intfNumber, altSetting);
  }
}
