package vasset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lakal3/vge/vge/vk"
	"io"
)

type DdsImageLoader struct {
}

func (d DdsImageLoader) SaveImage(ctx vk.APIContext, kind string, desc vk.ImageDescription, buffer *vk.Buffer) (content []byte) {
	return saveDds(ctx, desc, buffer)
}

func (d DdsImageLoader) SupportsImage(kind string) (read bool, write bool) {
	if kind == "dds" {
		return true, true
	}
	return false, false
}

func (d DdsImageLoader) DescribeImage(ctx vk.APIContext, kind string, desc *vk.ImageDescription, content []byte) {
	describeDds(ctx, desc, content)
}

func (d DdsImageLoader) LoadImage(ctx vk.APIContext, kind string, content []byte, buffer *vk.Buffer) {
	loadDds(ctx, content, buffer)
}

const ddsMagic = uint32(0x20534444)
const ddsDX10 = uint32(0x30315844)

const ddsHeaderLength = 4 + 124 + 20

type ddsHeader [ddsHeaderLength]byte

func (h *ddsHeader) getMagic() uint32 {
	return binary.LittleEndian.Uint32(h[:4])
}

func (h *ddsHeader) getFlags() uint32 {
	return binary.LittleEndian.Uint32(h[8:12])
}

func (h *ddsHeader) getSize(description *vk.ImageDescription) {
	description.Height, description.Width, description.Depth = binary.LittleEndian.Uint32(h[12:16]),
		binary.LittleEndian.Uint32(h[16:20]), binary.LittleEndian.Uint32(h[24:28])
	if (h.getFlags() & 0x800000) == 0 {
		description.Depth = 1
	}
}

func (h *ddsHeader) setSize(description vk.ImageDescription) {
	binary.LittleEndian.PutUint32(h[12:16], description.Height)
	binary.LittleEndian.PutUint32(h[16:20], description.Width)
	binary.LittleEndian.PutUint32(h[24:28], description.Depth)
}

func (h *ddsHeader) getMidmaps() uint32 {
	if (h.getFlags() & 0x20000) == 0 {
		return 1
	}
	return binary.LittleEndian.Uint32(h[28:32])
}

func (h *ddsHeader) setMidmaps(mps uint32) {
	binary.LittleEndian.PutUint32(h[28:32], mps)
}

func (h *ddsHeader) checkDDS10() bool {
	if binary.LittleEndian.Uint32(h[84:88]) != ddsDX10 {
		return false
	}
	return true
}

func (h *ddsHeader) getFormat() vk.Format {
	f := binary.LittleEndian.Uint32(h[128:132])
	for _, fi := range vk.Formats {
		if fi.DXGIFormat == f {
			return fi.Format
		}
	}
	return vk.FORMATUndefined
}

func (h *ddsHeader) setFormat(format vk.Format) error {
	fi, ok := vk.Formats[format]
	if !ok && fi.DXGIFormat == 0 {
		return fmt.Errorf("Unknown or unsupported format %v", format)
	}
	binary.LittleEndian.PutUint32(h[128:132], uint32(fi.DXGIFormat))
	return nil
}

func (h *ddsHeader) getLayers() uint32 {
	return binary.LittleEndian.Uint32(h[140:144])
}

func (h *ddsHeader) setLayers(layers uint32) {
	binary.LittleEndian.PutUint32(h[140:144], layers)
}

func newDdsHeader() ddsHeader {
	var h ddsHeader
	binary.LittleEndian.PutUint32(h[:4], ddsMagic)
	binary.LittleEndian.PutUint32(h[4:8], 124) // Size
	// Flags width, height, format, midmap size, depth
	binary.LittleEndian.PutUint32(h[8:12], 0x2|0x4|0x1000|0x20000|0x800000)
	// Format size
	binary.LittleEndian.PutUint32(h[76:80], 32)
	// Format flags
	binary.LittleEndian.PutUint32(h[80:84], 0x00000004)
	// DX10 for extended header
	binary.LittleEndian.PutUint32(h[84:88], ddsDX10)
	// 2D image
	binary.LittleEndian.PutUint32(h[132:136], 3)
	// Alpha mode straigh
	binary.LittleEndian.PutUint32(h[144:148], 1)
	return h
}

func describeDds(ctx vk.APIContext, desc *vk.ImageDescription, content []byte) {
	if !ctx.IsValid() {
		return
	}
	if len(content) < ddsHeaderLength {
		ctx.SetError(errors.New("Header too short"))
		return
	}
	var header ddsHeader
	copy(header[:], content)
	if !header.checkDDS10() {
		ctx.SetError(errors.New("Only new DX10 format supported"))
		return
	}
	header.getSize(desc)
	desc.MipLevels = header.getMidmaps()
	desc.Layers = header.getLayers()
	desc.Format = header.getFormat()

	if desc.Format == vk.FORMATUndefined {
		ctx.SetError(fmt.Errorf("Unknown DDS format %d", binary.LittleEndian.Uint32(header[128:132])))
		return
	}
	return
}

func loadDds(ctx vk.APIContext, content []byte, buffer *vk.Buffer) {
	buffer.IsValid(ctx)
	if !ctx.IsValid() {
		return
	}
	raw := content[ddsHeaderLength:]
	if uint64(len(raw)) > buffer.Size {
		ctx.SetError(fmt.Errorf("Buffer length should be %d not %d", len(raw), buffer.Size))
	}
	copy(buffer.Bytes(ctx), raw)
}

func saveDds(ctx vk.APIContext, desc vk.ImageDescription, buffer *vk.Buffer) []byte {
	if !buffer.IsValid(ctx) || !ctx.IsValid() {
		return nil
	}
	content := buffer.Bytes(ctx)
	w := &bytes.Buffer{}
	err := WriteDDS(w, desc, content)
	if err != nil {
		ctx.SetError(err)
	}
	return w.Bytes()
}

func WriteDDS(writer io.Writer, desc vk.ImageDescription, bytes []byte) error {
	h := newDdsHeader()
	h.setSize(desc)
	h.setLayers(desc.Layers)
	h.setMidmaps(desc.MipLevels)
	err := h.setFormat(desc.Format)
	if err != nil {
		return err
	}
	_, err = writer.Write(h[:])
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	return err
}
