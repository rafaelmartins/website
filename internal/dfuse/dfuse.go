package dfuse

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

/*
 * implementation based on information from:
 * UM0391 - DfuSe File Format Specification
 *
 * and reverse engineering on actual DfuSe files
 */

type DfuSeElement struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
	Data  []byte `json:"data"`
}

type DfuSeTarget struct {
	Name             string         `json:"name"`
	AlternateSetting byte           `json:"alternateSetting"`
	Elements         []DfuSeElement `json:"elements"`
}

type DfuSe struct {
	IdProduct uint16 `json:"idProduct"`
	IdVendor  uint16 `json:"idVendor"`

	Targets []DfuSeTarget `json:"targets"`
}

func NewFromArchive(f string, r io.ReadCloser) (*DfuSe, error) {
	dec, err := decompress(f, r)
	if err != nil {
		return nil, err
	}
	return NewFromReader(dec)
}

func NewFromReader(r io.Reader) (*DfuSe, error) {
	prefix := struct {
		Signature [5]byte
		Version   byte
		ImageSize uint32
		Targets   byte
	}{}
	if err := binary.Read(r, binary.LittleEndian, &prefix); err != nil {
		return nil, err
	}

	if !bytes.Equal(prefix.Signature[:], []byte("DfuSe")) {
		return nil, fmt.Errorf("dfuse: invalid prefix signature: %s", prefix.Signature[:])
	}
	if prefix.Version != 1 {
		return nil, fmt.Errorf("dfuse: unsupported version: %d", prefix.Version)
	}

	rv := &DfuSe{}

	for range prefix.Targets {
		target := struct {
			Signature        [6]byte
			AlternateSetting byte
			TargetNamed      byte
			Padding          [3]byte
			TargetName       [255]byte
			TargetSize       uint32
			NbElements       uint32
		}{}
		if err := binary.Read(r, binary.LittleEndian, &target); err != nil {
			return nil, err
		}

		if !bytes.Equal(target.Signature[:], []byte("Target")) {
			return nil, fmt.Errorf("dfuse: invalid target signature: %s", target.Signature[:])
		}

		tgt := DfuSeTarget{
			AlternateSetting: target.AlternateSetting,
		}
		if target.TargetNamed != 0 {
			if idx := bytes.IndexByte(target.TargetName[:], 0); idx >= 0 {
				tgt.Name = string(target.TargetName[:idx])
			}
		}

		for range target.NbElements {
			element := struct {
				ElementAddress uint32
				ElementSize    uint32
			}{}
			if err := binary.Read(r, binary.LittleEndian, &element); err != nil {
				return nil, err
			}

			buf := make([]byte, element.ElementSize)
			if _, err := r.Read(buf); err != nil {
				return nil, err
			}
			tgt.Elements = append(tgt.Elements, DfuSeElement{
				Start: element.ElementAddress,
				End:   element.ElementAddress + uint32(len(buf)),
				Data:  buf,
			})
		}

		rv.Targets = append(rv.Targets, tgt)
	}

	suffix := struct {
		BcdDevice    uint16
		IdProduct    uint16
		IdVendor     uint16
		BcdDfu       uint16
		DfuSignature [3]byte
		Length       byte
		Crc          uint32
	}{}
	if err := binary.Read(r, binary.LittleEndian, &suffix); err != nil {
		return nil, err
	}

	if !bytes.Equal(suffix.DfuSignature[:], []byte("UFD")) {
		return nil, fmt.Errorf("dfuse: invalid suffix signature: %s", suffix.DfuSignature[:])
	}
	if suffix.Length != 16 {
		return nil, fmt.Errorf("dfuse: invalid suffix length: %d", suffix.Length)
	}
	if suffix.BcdDfu != 0x011a {
		return nil, fmt.Errorf("dfuse: invalid dfu specification number: %#x", suffix.BcdDfu)
	}

	rv.IdProduct = suffix.IdProduct
	rv.IdVendor = suffix.IdVendor
	return rv, nil
}

func (d *DfuSe) ToJson() (io.ReadCloser, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(d); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}
