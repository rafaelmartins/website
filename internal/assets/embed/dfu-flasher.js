import { DfuSe } from "./dfuse.js";

const dfuJs = document.getElementById("dfu-js");
const dfuSupported = document.getElementById("dfu-supported");
const dfuProjectSelect = document.getElementById("dfu-project-select");
const dfuFirmwareSelect = document.getElementById("dfu-firmware-select");
const dfuConnectionBtn = document.getElementById("dfu-connection-btn");
const dfuConnectionIcon = document.getElementById("dfu-connection-icon");
const dfuConnectionBody = document.getElementById("dfu-connection-body");
const dfuFlashBtn = document.getElementById("dfu-flash-btn");
const dfuStatus = document.getElementById("dfu-status");
const dfuStatusLabel = document.getElementById("dfu-status-label");
const dfuStatusProgress = document.getElementById("dfu-status-progress");
const dfuLog = document.getElementById("dfu-log");
const dfuLogBody = document.getElementById("dfu-log-body");

let dfu = null;
let currentProject = null;
let projects = [];

function writeAndShow(div, msg) {
  const d = document.getElementById(div);
  const b = document.getElementById(`${div}-body`);
  if (!d || !b)
    return;

  b.textContent = msg;
  d.style.display = "";
}

function eraseAndHide(div) {
  const d = document.getElementById(div);
  const b = document.getElementById(`${div}-body`);
  if (!d || !b)
    return;

  d.style.display = "none";
  b.textContent = "";
}

function log(msg) {
  dfuLog.style.display = "";
  dfuLogBody.textContent += `${msg}\n`;
  dfuLogBody.scrollTop = dfuLog.scrollHeight;
}

const phaseLabels = { erase: "Memory Erase", write: "Memory Write" };

function progress(phase, sent, total) {
  const label = phaseLabels[phase] ?? "Status";

  dfuStatus.style.display = "";
  dfuStatusLabel.textContent = label;
  dfuStatusProgress.value = sent;
  dfuStatusProgress.max = total;
}

function btnConnection(connected, disabled) {
  dfuConnectionBody.textContent = connected ? "Disconnect" : "Connect";
  dfuConnectionIcon.className = `fa-solid ${connected ? "fa-plug-circle-xmark" : "fa-plug"}`;
  dfuConnectionBtn.dataset.connected = connected;
  dfuConnectionBtn.disabled = disabled;
}

function setError(msg) {
  if (!(msg instanceof Error)) {
    writeAndShow("dfu-error", msg);
    return;
  }

  console.log(msg);
  writeAndShow("dfu-error", msg.message);
}

function resetError() {
  eraseAndHide("dfu-error");
}

async function getProjects() {
  const projectsJson = dfuJs.dataset?.json?.split(";");
  if (projectsJson === null)
    throw new Error("no projects found");

  return Promise.all(projectsJson.map(async (url) => {
    const resp = await fetch(url);
    const result = await resp.json();
    for (const fw of result.firmwares) {
      const fwResp = await fetch(fw.url);
      fw.data = await fwResp.json();
    }
    return result;
  }));
}

function disconnect() {
  currentProject = null;

  dfuProjectSelect.disabled = false;
  if (dfuProjectSelect.value === "") {
    dfuFirmwareSelect.disabled = true;
  } else if (dfuFirmwareSelect.value !== "") {
    dfuFirmwareSelect.disabled = false;
    btnConnection(false, false);
    dfuFlashBtn.style.display = "none";
  }
}

try {
  dfu = new DfuSe();
  if (!Uint8Array.fromBase64)
    setError("Your browser does not support recent features required by this software. Please upgrade!");
  else {
    dfuSupported.style.display = "";
    projects = await getProjects();
  }
}
catch (e) {
  setError(e);
}

dfuProjectSelect.addEventListener("change", (event) => {
  resetError();

  eraseAndHide("dfu-firmwareinfo");
  btnConnection(false, true);

  for (let i = dfuFirmwareSelect.length - 1; i > 0; i--)
    dfuFirmwareSelect.remove(i);

  if (event.target.value === "") {
    dfuFirmwareSelect.disabled = true;
    return;
  }

  const prj = projects[event.target.value];
  if (!prj?.firmwares)
    return;

  prj.firmwares.forEach((fw, i) => {
    const opt = document.createElement("option");
    opt.value = i;
    opt.textContent = `${fw.name} - ${fw.version} - (${fw.type})`;
    dfuFirmwareSelect.appendChild(opt);
    dfuFirmwareSelect.disabled = false;
  });
});

dfuFirmwareSelect.addEventListener("change", (event) => {
  resetError();

  if (dfuFirmwareSelect.value === "") {
    eraseAndHide("dfu-firmwareinfo");
    return;
  }

  const fw = projects[dfuProjectSelect.value]?.firmwares[dfuFirmwareSelect.value];
  if (!fw)
    return;

  const size = fw.data.targets
    .flatMap((tgt) => tgt.elements)
    .reduce((sum, el) => sum + el.end - el.start, 0);

  writeAndShow("dfu-firmwareinfo", `Firmware size: ${size} bytes`);
  btnConnection(false, event.target.value === "");
});

dfuConnectionBtn.addEventListener("click", async () => {
  resetError();

  eraseAndHide("dfu-deviceinfo");

  if (dfuConnectionBtn.dataset.connected === "true") {
    disconnect();
    return;
  }

  dfuStatus.style.display = "none";
  dfuLogBody.textContent = "";
  dfuLog.style.display = "none";

  currentProject = projects[dfuProjectSelect.value]?.firmwares[dfuFirmwareSelect.value];
  if (!currentProject)
    return;

  try {
    const alternates = await dfu.connect(
      () => disconnect(),
      currentProject?.data?.idVendor || 0,
      currentProject?.data?.idProduct || 0,
    );

    const altLines = alternates.map((alt) => `${alt.name}: ${alt.totalSize} bytes`);
    const p = dfu.properties;
    const msg = [
      `Manufacturer: ${dfu.device.manufacturerName} (0x${dfu.device.vendorId.toString(16).padStart(4, "0")})`,
      `Product: ${dfu.device.productName} (0x${dfu.device.productId.toString(16).padStart(4, "0")})`,
      `Serial Number: ${dfu.device.serialNumber}`,
      "",
      ...altLines,
      "",
      "Properties:",
      `- Will Detach: ${p.WillDetach}`,
      `- Manifestation Tolerant: ${p.ManifestationTolerant}`,
      `- Can Upload: ${p.CanUpload}`,
      `- Can Download: ${p.CanDownload}`,
      `- Transfer Size: ${p.TransferSize}`,
      `- Detach TimeOut: ${p.DetachTimeOut}`,
      `- DFU Version: 0x${p.DFUVersion.toString(16).padStart(4, "0")}`,
    ].join("\n");

    writeAndShow("dfu-deviceinfo", msg);
    dfuProjectSelect.disabled = true;
    dfuFirmwareSelect.disabled = true;
    btnConnection(true, false);
    dfuFlashBtn.style.display = "";
  } catch (e) {
    disconnect();
    if (e.name === "NotFoundError") {
      setError("No device selected.");
      return;
    }
    setError(e);
  }
});

dfuFlashBtn.addEventListener("click", async () => {
  resetError();

  if (!currentProject?.data?.targets) {
    setError("No firmware loaded!");
    return;
  }

  dfuStatus.style.display = "";

  try {
    for (const target of currentProject.data.targets) {
      await dfu.configure(target.alternateSetting);

      for (const element of target.elements)
        await dfu.write(Uint8Array.fromBase64(element.data), element.start, progress, log);
    }
  }
  catch (e) {
    disconnect();
    setError(e);
  }
});

projects.forEach((prj, i) => {
  const opt = document.createElement("option");
  opt.value = i;
  opt.textContent = prj.name;
  dfuProjectSelect.appendChild(opt);
  dfuProjectSelect.disabled = false;
});
