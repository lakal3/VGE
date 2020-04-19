package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type enumInfo struct {
	name   string
	prefix string
}

var allEnums []enumInfo = []enumInfo{
	{name: "VkBufferUsageFlagBits"},
	{name: "VkFormat"},
	{name: "VkQueueFlagBits"},
	{name: "VkDescriptorType"},
	{name: "VkSamplerAddressMode"},
	{name: "VkImageUsageFlagBits"},
	{name: "VkImageLayout"},
	{name: "VkImageAspectFlagBits"},
	{name: "VkShaderStageFlagBits"},
	{name: "VkVertexInputRate"},
	{name: "VkPipelineStageFlagBits"},
	{name: "VkPhysicalDeviceType"},
	{name: "VkDescriptorBindingFlagBitsEXT"},
}

/*
   new EnumInfo() { EnumName = "VkImageUsageFlags", Prefix = "VK_IMAGE_USAGE_", Flags = true},
   new EnumInfo() { EnumName = "VkFormat", Prefix = "VK_FORMAT_", End = "VK_FORMAT_G8B8G8R8_422_UNORM_KHR", Flags = false},
   new EnumInfo() { EnumName = "VkQueueFlagBits", Prefix = "VK_QUEUE_", Start = "VK_QUEUE_GRAPHICS_BIT",  End = "VK_QUEUE_FLAG_BITS_MAX_ENUM", Flags = true},
   new EnumInfo() { EnumName = "VkDescriptorType", Prefix = "VK_DESCRIPTOR_TYPE_",   End = "VK_DESCRIPTOR_TYPE_BEGIN_RANGE", Flags = false},
   new EnumInfo() { EnumName = "VkSamplerAddressMode", Prefix = "VK_SAMPLER_ADDRESS_MODE_",   End = "VK_SAMPLER_ADDRESS_MODE_MIRROR_CLAMP_TO_EDGE_KHR", Flags = false},

*/

func genGoEnums(writer io.Writer) error {
	vdsk := os.Getenv("VULKAN_SDK")
	if len(vdsk) == 0 {
		return errors.New("No VULKAN_SDK enviroment variable")
	}
	incPath := filepath.Join(vdsk, "Include/vulkan/vulkan_core.h")
	fInc, err := os.Open(incPath)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %v", incPath, err)
	}
	defer fInc.Close()

	_, err = fmt.Fprintln(writer, "package vk")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(writer, "")
	sc := bufio.NewScanner(fInc)
	for sc.Scan() {
		line := sc.Text()
		if strings.Index(line, "enum") > 0 {
			err = parseEnum(writer, sc, line)
			if err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

func parseEnum(writer io.Writer, sc *bufio.Scanner, line string) (err error) {
	for _, en := range allEnums {
		if strings.Index(line, en.name+" ") > 0 {
			err = emitEnum(writer, sc, en)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func emitEnum(writer io.Writer, sc *bufio.Scanner, info enumInfo) error {
	eName := goEnumName(info)
	bits := strings.HasSuffix(eName, "Bits")
	if bits {
		eName = eName[:len(eName)-4] + "s"
	}
	_, _ = fmt.Fprintln(writer, "type ", eName, " int32")
	_, _ = fmt.Fprintln(writer, "const (")
	for sc.Scan() {
		l := sc.Text()
		if strings.IndexRune(l, '}') >= 0 {
			break
		}
		if strings.HasSuffix(l, ",") {
			l = l[:len(l)-1]
		}
		idxEq := strings.IndexRune(l, '=')
		idxPre := strings.Index(l, info.prefix)
		if idxEq < 0 || idxPre < 0 {
			continue
		}
		if strings.Index(l, "MAX_ENUM") > 0 {
			continue
		}

		val, err := strconv.ParseInt(strings.Trim(l[idxEq+1:], " \t"), 0, 32)
		if err != nil {
			continue
		}
		if bits {
			_, _ = fmt.Fprintf(writer, "    %s = %s(0x%x)\n", goFlagName(l[idxPre+len(info.prefix):idxEq-1]), eName, val)
		} else {
			_, _ = fmt.Fprintf(writer, "    %s = %s(%d)\n", goFlagName(l[idxPre+len(info.prefix):idxEq-1]), eName, val)
		}
	}
	_, _ = fmt.Fprintln(writer, ")")
	_, _ = fmt.Fprintln(writer, "")
	return sc.Err()
}

func goFlagName(cName string) string {
	cName = strings.Trim(cName, " \t")
	cap := true
	segment := 0
	sb := strings.Builder{}
	for _, r := range cName {
		if r == '_' {
			cap = true
			segment++
		} else {
			if segment == 0 {
				continue
			}
			if cap {
				sb.WriteRune(unicode.ToUpper(r))
				if segment > 1 {
					cap = false
				}
			} else {
				sb.WriteRune(unicode.ToLower(r))
			}
		}
	}
	return sb.String()
}

func goEnumName(info enumInfo) string {
	return info.name[2:]
}
