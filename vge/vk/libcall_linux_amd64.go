package vk

// Autogenerated file - do not edit!

import (
	"github.com/lakal3/vge/vge/dldyn"
	"unsafe"
)

var libcall struct {
	h_lib dldyn.Handle

	t_AddDynamicDescriptors             uintptr
	t_AddValidation                     uintptr
	t_AddValidationException            uintptr
	t_Application_Init                  uintptr
	t_Buffer_CopyFrom                   uintptr
	t_Buffer_GetPtr                     uintptr
	t_Buffer_NewView                    uintptr
	t_Command_Begin                     uintptr
	t_Command_BeginRenderPass           uintptr
	t_Command_ClearImage                uintptr
	t_Command_Compute                   uintptr
	t_Command_CopyBuffer                uintptr
	t_Command_CopyBufferToImage         uintptr
	t_Command_CopyImageToBuffer         uintptr
	t_Command_Draw                      uintptr
	t_Command_EndRenderPass             uintptr
	t_Command_SetLayout                 uintptr
	t_Command_Wait                      uintptr
	t_Command_WriteTimer                uintptr
	t_ComputePipeline_Create            uintptr
	t_DebugPoint                        uintptr
	t_DescriptorLayout_NewPool          uintptr
	t_DescriptorPool_Alloc              uintptr
	t_DescriptorSet_WriteBuffer         uintptr
	t_DescriptorSet_WriteBufferView     uintptr
	t_DescriptorSet_WriteImage          uintptr
	t_Desktop_CreateWindow              uintptr
	t_Desktop_GetKeyName                uintptr
	t_Desktop_GetMonitor                uintptr
	t_Desktop_PullEvent                 uintptr
	t_Device_NewBuffer                  uintptr
	t_Device_NewCommand                 uintptr
	t_Device_NewComputePipeline         uintptr
	t_Device_NewDescriptorLayout        uintptr
	t_Device_NewGraphicsPipeline        uintptr
	t_Device_NewImage                   uintptr
	t_Device_NewMemoryBlock             uintptr
	t_Device_NewSampler                 uintptr
	t_Device_NewTimestampQuery          uintptr
	t_Device_Submit                     uintptr
	t_Disposable_Dispose                uintptr
	t_Exception_GetError                uintptr
	t_GraphicsPipeline_AddAlphaBlend    uintptr
	t_GraphicsPipeline_AddDepth         uintptr
	t_GraphicsPipeline_AddVertexBinding uintptr
	t_GraphicsPipeline_AddVertexFormat  uintptr
	t_GraphicsPipeline_Create           uintptr
	t_GraphicsPipeline_SetTopology      uintptr
	t_ImageLoader_Describe              uintptr
	t_ImageLoader_Load                  uintptr
	t_ImageLoader_Save                  uintptr
	t_ImageLoader_Supported             uintptr
	t_Image_NewView                     uintptr
	t_Instance_GetPhysicalDevice        uintptr
	t_Instance_NewDevice                uintptr
	t_MemoryBlock_Allocate              uintptr
	t_MemoryBlock_Reserve               uintptr
	t_NewApplication                    uintptr
	t_NewDesktop                        uintptr
	t_NewImageLoader                    uintptr
	t_NewRenderPass                     uintptr
	t_Pipeline_AddDescriptorLayout      uintptr
	t_Pipeline_AddShader                uintptr
	t_QueryPool_Get                     uintptr
	t_RenderPass_NewFrameBuffer         uintptr
	t_RenderPass_NewNullFrameBuffer     uintptr
	t_Window_GetNextFrame               uintptr
	t_Window_GetPos                     uintptr
	t_Window_PrepareSwapchain           uintptr
	t_Window_SetPos                     uintptr
}

func loadLib() (err error) {
	if libcall.h_lib != 0 {
		return nil
	}
	libcall.h_lib, err = dldyn.DLOpen(GetDllPath())
	if err != nil {
		return err
	}

	libcall.t_AddDynamicDescriptors, err = dldyn.GetProcAddress(libcall.h_lib, "AddDynamicDescriptors")
	if err != nil {
		return err
	}
	libcall.t_AddValidation, err = dldyn.GetProcAddress(libcall.h_lib, "AddValidation")
	if err != nil {
		return err
	}
	libcall.t_AddValidationException, err = dldyn.GetProcAddress(libcall.h_lib, "AddValidationException")
	if err != nil {
		return err
	}
	libcall.t_Application_Init, err = dldyn.GetProcAddress(libcall.h_lib, "Application_Init")
	if err != nil {
		return err
	}
	libcall.t_Buffer_CopyFrom, err = dldyn.GetProcAddress(libcall.h_lib, "Buffer_CopyFrom")
	if err != nil {
		return err
	}
	libcall.t_Buffer_GetPtr, err = dldyn.GetProcAddress(libcall.h_lib, "Buffer_GetPtr")
	if err != nil {
		return err
	}
	libcall.t_Buffer_NewView, err = dldyn.GetProcAddress(libcall.h_lib, "Buffer_NewView")
	if err != nil {
		return err
	}
	libcall.t_Command_Begin, err = dldyn.GetProcAddress(libcall.h_lib, "Command_Begin")
	if err != nil {
		return err
	}
	libcall.t_Command_BeginRenderPass, err = dldyn.GetProcAddress(libcall.h_lib, "Command_BeginRenderPass")
	if err != nil {
		return err
	}
	libcall.t_Command_ClearImage, err = dldyn.GetProcAddress(libcall.h_lib, "Command_ClearImage")
	if err != nil {
		return err
	}
	libcall.t_Command_Compute, err = dldyn.GetProcAddress(libcall.h_lib, "Command_Compute")
	if err != nil {
		return err
	}
	libcall.t_Command_CopyBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "Command_CopyBuffer")
	if err != nil {
		return err
	}
	libcall.t_Command_CopyBufferToImage, err = dldyn.GetProcAddress(libcall.h_lib, "Command_CopyBufferToImage")
	if err != nil {
		return err
	}
	libcall.t_Command_CopyImageToBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "Command_CopyImageToBuffer")
	if err != nil {
		return err
	}
	libcall.t_Command_Draw, err = dldyn.GetProcAddress(libcall.h_lib, "Command_Draw")
	if err != nil {
		return err
	}
	libcall.t_Command_EndRenderPass, err = dldyn.GetProcAddress(libcall.h_lib, "Command_EndRenderPass")
	if err != nil {
		return err
	}
	libcall.t_Command_SetLayout, err = dldyn.GetProcAddress(libcall.h_lib, "Command_SetLayout")
	if err != nil {
		return err
	}
	libcall.t_Command_Wait, err = dldyn.GetProcAddress(libcall.h_lib, "Command_Wait")
	if err != nil {
		return err
	}
	libcall.t_Command_WriteTimer, err = dldyn.GetProcAddress(libcall.h_lib, "Command_WriteTimer")
	if err != nil {
		return err
	}
	libcall.t_ComputePipeline_Create, err = dldyn.GetProcAddress(libcall.h_lib, "ComputePipeline_Create")
	if err != nil {
		return err
	}
	libcall.t_DebugPoint, err = dldyn.GetProcAddress(libcall.h_lib, "DebugPoint")
	if err != nil {
		return err
	}
	libcall.t_DescriptorLayout_NewPool, err = dldyn.GetProcAddress(libcall.h_lib, "DescriptorLayout_NewPool")
	if err != nil {
		return err
	}
	libcall.t_DescriptorPool_Alloc, err = dldyn.GetProcAddress(libcall.h_lib, "DescriptorPool_Alloc")
	if err != nil {
		return err
	}
	libcall.t_DescriptorSet_WriteBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteBuffer")
	if err != nil {
		return err
	}
	libcall.t_DescriptorSet_WriteBufferView, err = dldyn.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteBufferView")
	if err != nil {
		return err
	}
	libcall.t_DescriptorSet_WriteImage, err = dldyn.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteImage")
	if err != nil {
		return err
	}
	libcall.t_Desktop_CreateWindow, err = dldyn.GetProcAddress(libcall.h_lib, "Desktop_CreateWindow")
	if err != nil {
		return err
	}
	libcall.t_Desktop_GetKeyName, err = dldyn.GetProcAddress(libcall.h_lib, "Desktop_GetKeyName")
	if err != nil {
		return err
	}
	libcall.t_Desktop_GetMonitor, err = dldyn.GetProcAddress(libcall.h_lib, "Desktop_GetMonitor")
	if err != nil {
		return err
	}
	libcall.t_Desktop_PullEvent, err = dldyn.GetProcAddress(libcall.h_lib, "Desktop_PullEvent")
	if err != nil {
		return err
	}
	libcall.t_Device_NewBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewBuffer")
	if err != nil {
		return err
	}
	libcall.t_Device_NewCommand, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewCommand")
	if err != nil {
		return err
	}
	libcall.t_Device_NewComputePipeline, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewComputePipeline")
	if err != nil {
		return err
	}
	libcall.t_Device_NewDescriptorLayout, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewDescriptorLayout")
	if err != nil {
		return err
	}
	libcall.t_Device_NewGraphicsPipeline, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewGraphicsPipeline")
	if err != nil {
		return err
	}
	libcall.t_Device_NewImage, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewImage")
	if err != nil {
		return err
	}
	libcall.t_Device_NewMemoryBlock, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewMemoryBlock")
	if err != nil {
		return err
	}
	libcall.t_Device_NewSampler, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewSampler")
	if err != nil {
		return err
	}
	libcall.t_Device_NewTimestampQuery, err = dldyn.GetProcAddress(libcall.h_lib, "Device_NewTimestampQuery")
	if err != nil {
		return err
	}
	libcall.t_Device_Submit, err = dldyn.GetProcAddress(libcall.h_lib, "Device_Submit")
	if err != nil {
		return err
	}
	libcall.t_Disposable_Dispose, err = dldyn.GetProcAddress(libcall.h_lib, "Disposable_Dispose")
	if err != nil {
		return err
	}
	libcall.t_Exception_GetError, err = dldyn.GetProcAddress(libcall.h_lib, "Exception_GetError")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_AddAlphaBlend, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddAlphaBlend")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_AddDepth, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddDepth")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_AddVertexBinding, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddVertexBinding")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_AddVertexFormat, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddVertexFormat")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_Create, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_Create")
	if err != nil {
		return err
	}
	libcall.t_GraphicsPipeline_SetTopology, err = dldyn.GetProcAddress(libcall.h_lib, "GraphicsPipeline_SetTopology")
	if err != nil {
		return err
	}
	libcall.t_ImageLoader_Describe, err = dldyn.GetProcAddress(libcall.h_lib, "ImageLoader_Describe")
	if err != nil {
		return err
	}
	libcall.t_ImageLoader_Load, err = dldyn.GetProcAddress(libcall.h_lib, "ImageLoader_Load")
	if err != nil {
		return err
	}
	libcall.t_ImageLoader_Save, err = dldyn.GetProcAddress(libcall.h_lib, "ImageLoader_Save")
	if err != nil {
		return err
	}
	libcall.t_ImageLoader_Supported, err = dldyn.GetProcAddress(libcall.h_lib, "ImageLoader_Supported")
	if err != nil {
		return err
	}
	libcall.t_Image_NewView, err = dldyn.GetProcAddress(libcall.h_lib, "Image_NewView")
	if err != nil {
		return err
	}
	libcall.t_Instance_GetPhysicalDevice, err = dldyn.GetProcAddress(libcall.h_lib, "Instance_GetPhysicalDevice")
	if err != nil {
		return err
	}
	libcall.t_Instance_NewDevice, err = dldyn.GetProcAddress(libcall.h_lib, "Instance_NewDevice")
	if err != nil {
		return err
	}
	libcall.t_MemoryBlock_Allocate, err = dldyn.GetProcAddress(libcall.h_lib, "MemoryBlock_Allocate")
	if err != nil {
		return err
	}
	libcall.t_MemoryBlock_Reserve, err = dldyn.GetProcAddress(libcall.h_lib, "MemoryBlock_Reserve")
	if err != nil {
		return err
	}
	libcall.t_NewApplication, err = dldyn.GetProcAddress(libcall.h_lib, "NewApplication")
	if err != nil {
		return err
	}
	libcall.t_NewDesktop, err = dldyn.GetProcAddress(libcall.h_lib, "NewDesktop")
	if err != nil {
		return err
	}
	libcall.t_NewImageLoader, err = dldyn.GetProcAddress(libcall.h_lib, "NewImageLoader")
	if err != nil {
		return err
	}
	libcall.t_NewRenderPass, err = dldyn.GetProcAddress(libcall.h_lib, "NewRenderPass")
	if err != nil {
		return err
	}
	libcall.t_Pipeline_AddDescriptorLayout, err = dldyn.GetProcAddress(libcall.h_lib, "Pipeline_AddDescriptorLayout")
	if err != nil {
		return err
	}
	libcall.t_Pipeline_AddShader, err = dldyn.GetProcAddress(libcall.h_lib, "Pipeline_AddShader")
	if err != nil {
		return err
	}
	libcall.t_QueryPool_Get, err = dldyn.GetProcAddress(libcall.h_lib, "QueryPool_Get")
	if err != nil {
		return err
	}
	libcall.t_RenderPass_NewFrameBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "RenderPass_NewFrameBuffer")
	if err != nil {
		return err
	}
	libcall.t_RenderPass_NewNullFrameBuffer, err = dldyn.GetProcAddress(libcall.h_lib, "RenderPass_NewNullFrameBuffer")
	if err != nil {
		return err
	}
	libcall.t_Window_GetNextFrame, err = dldyn.GetProcAddress(libcall.h_lib, "Window_GetNextFrame")
	if err != nil {
		return err
	}
	libcall.t_Window_GetPos, err = dldyn.GetProcAddress(libcall.h_lib, "Window_GetPos")
	if err != nil {
		return err
	}
	libcall.t_Window_PrepareSwapchain, err = dldyn.GetProcAddress(libcall.h_lib, "Window_PrepareSwapchain")
	if err != nil {
		return err
	}
	libcall.t_Window_SetPos, err = dldyn.GetProcAddress(libcall.h_lib, "Window_SetPos")
	if err != nil {
		return err
	}
	return nil
}

func call_AddDynamicDescriptors(ctx APIContext, app hApplication) {
	atEnd := ctx.Begin("AddDynamicDescriptors")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_AddDynamicDescriptors, 1, uintptr(app), 0, 0)
	handleError(ctx, rc)
}
func call_AddValidation(ctx APIContext, app hApplication) {
	atEnd := ctx.Begin("AddValidation")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_AddValidation, 1, uintptr(app), 0, 0)
	handleError(ctx, rc)
}
func call_AddValidationException(ctx APIContext, msgId int32) {
	atEnd := ctx.Begin("AddValidationException")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_AddValidationException, 1, uintptr(msgId), 0, 0)
	handleError(ctx, rc)
}
func call_Application_Init(ctx APIContext, app hApplication, inst *hInstance) {
	_tmp_inst := *inst
	atEnd := ctx.Begin("Application_Init")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Application_Init, 2, uintptr(app), uintptr(unsafe.Pointer(&_tmp_inst)), 0)
	handleError(ctx, rc)
	*inst = _tmp_inst
}
func call_Buffer_CopyFrom(ctx APIContext, buffer hBuffer, offset uint64, ptr uintptr, size uint64) {
	atEnd := ctx.Begin("Buffer_CopyFrom")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Buffer_CopyFrom, 4, uintptr(buffer), uintptr(offset), uintptr(ptr), uintptr(size), 0, 0)
	handleError(ctx, rc)
}
func call_Buffer_GetPtr(ctx APIContext, buffer hBuffer, ptr *uintptr) {
	_tmp_ptr := *ptr
	atEnd := ctx.Begin("Buffer_GetPtr")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Buffer_GetPtr, 2, uintptr(buffer), uintptr(unsafe.Pointer(&_tmp_ptr)), 0)
	handleError(ctx, rc)
	*ptr = _tmp_ptr
}
func call_Buffer_NewView(ctx APIContext, buffer hBuffer, format Format, offset uint64, size uint64, view *hBufferView) {
	_tmp_view := *view
	atEnd := ctx.Begin("Buffer_NewView")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Buffer_NewView, 5, uintptr(buffer), uintptr(format), uintptr(offset), uintptr(size), uintptr(unsafe.Pointer(&_tmp_view)), 0)
	handleError(ctx, rc)
	*view = _tmp_view
}
func call_Command_Begin(ctx APIContext, cmd hCommand) {
	atEnd := ctx.Begin("Command_Begin")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_Begin, 1, uintptr(cmd), 0, 0)
	handleError(ctx, rc)
}
func call_Command_BeginRenderPass(ctx APIContext, cmd hCommand, rp hRenderPass, fb hFramebuffer) {
	atEnd := ctx.Begin("Command_BeginRenderPass")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_BeginRenderPass, 3, uintptr(cmd), uintptr(rp), uintptr(fb))
	handleError(ctx, rc)
}
func call_Command_ClearImage(ctx APIContext, cmd hCommand, dst hImage, imRange *ImageRange, layout ImageLayout, color float32, alpha float32) {
	_tmp_imRange := *imRange
	atEnd := ctx.Begin("Command_ClearImage")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Command_ClearImage, 6, uintptr(cmd), uintptr(dst), uintptr(unsafe.Pointer(&_tmp_imRange)), uintptr(layout), uintptr(color), uintptr(alpha))
	handleError(ctx, rc)
	*imRange = _tmp_imRange
}
func call_Command_Compute(ctx APIContext, hCmd hCommand, hPl hComputePipeline, x uint32, y uint32, z uint32, descriptors []hDescriptorSet) {
	atEnd := ctx.Begin("Command_Compute")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke9(libcall.t_Command_Compute, 7, uintptr(hCmd), uintptr(hPl), uintptr(x), uintptr(y), uintptr(z), sliceToUintptr(descriptors), uintptr(len(descriptors)), 0, 0)
	handleError(ctx, rc)
}
func call_Command_CopyBuffer(ctx APIContext, cmd hCommand, src hBuffer, dst hBuffer) {
	atEnd := ctx.Begin("Command_CopyBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_CopyBuffer, 3, uintptr(cmd), uintptr(src), uintptr(dst))
	handleError(ctx, rc)
}
func call_Command_CopyBufferToImage(ctx APIContext, cmd hCommand, src hBuffer, dst hImage, imRange *ImageRange, offset uint64) {
	_tmp_imRange := *imRange
	atEnd := ctx.Begin("Command_CopyBufferToImage")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Command_CopyBufferToImage, 5, uintptr(cmd), uintptr(src), uintptr(dst), uintptr(unsafe.Pointer(&_tmp_imRange)), uintptr(offset), 0)
	handleError(ctx, rc)
	*imRange = _tmp_imRange
}
func call_Command_CopyImageToBuffer(ctx APIContext, cmd hCommand, src hImage, dst hBuffer, imRange *ImageRange, offset uint64) {
	_tmp_imRange := *imRange
	atEnd := ctx.Begin("Command_CopyImageToBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Command_CopyImageToBuffer, 5, uintptr(cmd), uintptr(src), uintptr(dst), uintptr(unsafe.Pointer(&_tmp_imRange)), uintptr(offset), 0)
	handleError(ctx, rc)
	*imRange = _tmp_imRange
}
func call_Command_Draw(ctx APIContext, cmd hCommand, draws []DrawItem) {
	atEnd := ctx.Begin("Command_Draw")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_Draw, 3, uintptr(cmd), sliceToUintptr(draws), uintptr(len(draws)))
	handleError(ctx, rc)
}
func call_Command_EndRenderPass(ctx APIContext, cmd hCommand) {
	atEnd := ctx.Begin("Command_EndRenderPass")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_EndRenderPass, 1, uintptr(cmd), 0, 0)
	handleError(ctx, rc)
}
func call_Command_SetLayout(ctx APIContext, cmd hCommand, image hImage, imRange *ImageRange, newLayout ImageLayout) {
	_tmp_imRange := *imRange
	atEnd := ctx.Begin("Command_SetLayout")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Command_SetLayout, 4, uintptr(cmd), uintptr(image), uintptr(unsafe.Pointer(&_tmp_imRange)), uintptr(newLayout), 0, 0)
	handleError(ctx, rc)
	*imRange = _tmp_imRange
}
func call_Command_Wait(ctx APIContext, cmd hCommand) {
	atEnd := ctx.Begin("Command_Wait")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Command_Wait, 1, uintptr(cmd), 0, 0)
	handleError(ctx, rc)
}
func call_Command_WriteTimer(ctx APIContext, cmd hCommand, qp hQueryPool, stages PipelineStageFlags, timerIndex uint32) {
	atEnd := ctx.Begin("Command_WriteTimer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Command_WriteTimer, 4, uintptr(cmd), uintptr(qp), uintptr(stages), uintptr(timerIndex), 0, 0)
	handleError(ctx, rc)
}
func call_ComputePipeline_Create(ctx APIContext, cp hComputePipeline) {
	atEnd := ctx.Begin("ComputePipeline_Create")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_ComputePipeline_Create, 1, uintptr(cp), 0, 0)
	handleError(ctx, rc)
}
func call_DebugPoint(point []byte) {
	_ = dldyn.Invoke(libcall.t_DebugPoint, 2, byteArrayToUintptr(point), uintptr(len(point)), 0)
}
func call_DescriptorLayout_NewPool(ctx APIContext, layout hDescriptorLayout, size uint32, pool *hDescriptorPool) {
	_tmp_pool := *pool
	atEnd := ctx.Begin("DescriptorLayout_NewPool")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_DescriptorLayout_NewPool, 3, uintptr(layout), uintptr(size), uintptr(unsafe.Pointer(&_tmp_pool)))
	handleError(ctx, rc)
	*pool = _tmp_pool
}
func call_DescriptorPool_Alloc(ctx APIContext, pool hDescriptorPool, ds *hDescriptorSet) {
	_tmp_ds := *ds
	atEnd := ctx.Begin("DescriptorPool_Alloc")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_DescriptorPool_Alloc, 2, uintptr(pool), uintptr(unsafe.Pointer(&_tmp_ds)), 0)
	handleError(ctx, rc)
	*ds = _tmp_ds
}
func call_DescriptorSet_WriteBuffer(ctx APIContext, ds hDescriptorSet, binding uint32, at uint32, buffer hBuffer, from uint64, size uint64) {
	atEnd := ctx.Begin("DescriptorSet_WriteBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_DescriptorSet_WriteBuffer, 6, uintptr(ds), uintptr(binding), uintptr(at), uintptr(buffer), uintptr(from), uintptr(size))
	handleError(ctx, rc)
}
func call_DescriptorSet_WriteBufferView(ctx APIContext, ds hDescriptorSet, binding uint32, at uint32, bufferView hBufferView) {
	atEnd := ctx.Begin("DescriptorSet_WriteBufferView")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_DescriptorSet_WriteBufferView, 4, uintptr(ds), uintptr(binding), uintptr(at), uintptr(bufferView), 0, 0)
	handleError(ctx, rc)
}
func call_DescriptorSet_WriteImage(ctx APIContext, ds hDescriptorSet, binding uint32, at uint32, view hImageView, sampler hSampler) {
	atEnd := ctx.Begin("DescriptorSet_WriteImage")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_DescriptorSet_WriteImage, 5, uintptr(ds), uintptr(binding), uintptr(at), uintptr(view), uintptr(sampler), 0)
	handleError(ctx, rc)
}
func call_Desktop_CreateWindow(ctx APIContext, desktop hDesktop, title []byte, pos *WindowPos, win *hWindow) {
	_tmp_pos := *pos
	_tmp_win := *win
	atEnd := ctx.Begin("Desktop_CreateWindow")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Desktop_CreateWindow, 5, uintptr(desktop), byteArrayToUintptr(title), uintptr(len(title)), uintptr(unsafe.Pointer(&_tmp_pos)), uintptr(unsafe.Pointer(&_tmp_win)), 0)
	handleError(ctx, rc)
	*pos = _tmp_pos
	*win = _tmp_win
}
func call_Desktop_GetKeyName(ctx APIContext, desktop hDesktop, keyCode uint32, name []uint8, strLen *uint32) {
	_tmp_strLen := *strLen
	atEnd := ctx.Begin("Desktop_GetKeyName")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Desktop_GetKeyName, 5, uintptr(desktop), uintptr(keyCode), sliceToUintptr(name), uintptr(len(name)), uintptr(unsafe.Pointer(&_tmp_strLen)), 0)
	handleError(ctx, rc)
	*strLen = _tmp_strLen
}
func call_Desktop_GetMonitor(ctx APIContext, desktop hDesktop, monitor uint32, info *WindowPos) {
	_tmp_info := *info
	atEnd := ctx.Begin("Desktop_GetMonitor")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Desktop_GetMonitor, 3, uintptr(desktop), uintptr(monitor), uintptr(unsafe.Pointer(&_tmp_info)))
	handleError(ctx, rc)
	*info = _tmp_info
}
func call_Desktop_PullEvent(ctx APIContext, desktop hDesktop, ev *RawEvent) {
	_tmp_ev := *ev
	atEnd := ctx.Begin("Desktop_PullEvent")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Desktop_PullEvent, 2, uintptr(desktop), uintptr(unsafe.Pointer(&_tmp_ev)), 0)
	handleError(ctx, rc)
	*ev = _tmp_ev
}
func call_Device_NewBuffer(ctx APIContext, dev hDevice, size uint64, hostMemory bool, usage BufferUsageFlags, buffer *hBuffer) {
	_tmp_buffer := *buffer
	atEnd := ctx.Begin("Device_NewBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Device_NewBuffer, 5, uintptr(dev), uintptr(size), boolToUintptr(hostMemory), uintptr(usage), uintptr(unsafe.Pointer(&_tmp_buffer)), 0)
	handleError(ctx, rc)
	*buffer = _tmp_buffer
}
func call_Device_NewCommand(ctx APIContext, dev hDevice, queueType QueueFlags, once bool, command *hCommand) {
	_tmp_command := *command
	atEnd := ctx.Begin("Device_NewCommand")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Device_NewCommand, 4, uintptr(dev), uintptr(queueType), boolToUintptr(once), uintptr(unsafe.Pointer(&_tmp_command)), 0, 0)
	handleError(ctx, rc)
	*command = _tmp_command
}
func call_Device_NewComputePipeline(ctx APIContext, dev hDevice, cp *hComputePipeline) {
	_tmp_cp := *cp
	atEnd := ctx.Begin("Device_NewComputePipeline")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Device_NewComputePipeline, 2, uintptr(dev), uintptr(unsafe.Pointer(&_tmp_cp)), 0)
	handleError(ctx, rc)
	*cp = _tmp_cp
}
func call_Device_NewDescriptorLayout(ctx APIContext, dev hDevice, descriptorType DescriptorType, stages ShaderStageFlags, element uint32, flags DescriptorBindingFlagBitsEXT, prevLayout hDescriptorLayout, dsLayout *hDescriptorLayout) {
	_tmp_dsLayout := *dsLayout
	atEnd := ctx.Begin("Device_NewDescriptorLayout")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke9(libcall.t_Device_NewDescriptorLayout, 7, uintptr(dev), uintptr(descriptorType), uintptr(stages), uintptr(element), uintptr(flags), uintptr(prevLayout), uintptr(unsafe.Pointer(&_tmp_dsLayout)), 0, 0)
	handleError(ctx, rc)
	*dsLayout = _tmp_dsLayout
}
func call_Device_NewGraphicsPipeline(ctx APIContext, dev hDevice, gp *hGraphicsPipeline) {
	_tmp_gp := *gp
	atEnd := ctx.Begin("Device_NewGraphicsPipeline")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Device_NewGraphicsPipeline, 2, uintptr(dev), uintptr(unsafe.Pointer(&_tmp_gp)), 0)
	handleError(ctx, rc)
	*gp = _tmp_gp
}
func call_Device_NewImage(ctx APIContext, dev hDevice, usage ImageUsageFlags, desc *ImageDescription, image *hImage) {
	_tmp_desc := *desc
	_tmp_image := *image
	atEnd := ctx.Begin("Device_NewImage")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Device_NewImage, 4, uintptr(dev), uintptr(usage), uintptr(unsafe.Pointer(&_tmp_desc)), uintptr(unsafe.Pointer(&_tmp_image)), 0, 0)
	handleError(ctx, rc)
	*desc = _tmp_desc
	*image = _tmp_image
}
func call_Device_NewMemoryBlock(ctx APIContext, dev hDevice, memBlock *hMemoryBlock) {
	_tmp_memBlock := *memBlock
	atEnd := ctx.Begin("Device_NewMemoryBlock")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Device_NewMemoryBlock, 2, uintptr(dev), uintptr(unsafe.Pointer(&_tmp_memBlock)), 0)
	handleError(ctx, rc)
	*memBlock = _tmp_memBlock
}
func call_Device_NewSampler(ctx APIContext, dev hDevice, repeatMode SamplerAddressMode, sampler *hSampler) {
	_tmp_sampler := *sampler
	atEnd := ctx.Begin("Device_NewSampler")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Device_NewSampler, 3, uintptr(dev), uintptr(repeatMode), uintptr(unsafe.Pointer(&_tmp_sampler)))
	handleError(ctx, rc)
	*sampler = _tmp_sampler
}
func call_Device_NewTimestampQuery(ctx APIContext, dev hDevice, size uint32, qp *hQueryPool) {
	_tmp_qp := *qp
	atEnd := ctx.Begin("Device_NewTimestampQuery")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Device_NewTimestampQuery, 3, uintptr(dev), uintptr(size), uintptr(unsafe.Pointer(&_tmp_qp)))
	handleError(ctx, rc)
	*qp = _tmp_qp
}
func call_Device_Submit(ctx APIContext, dev hDevice, cmd hCommand, priority uint32, info []hSubmitInfo, waitStage PipelineStageFlags, waitInfo *hSubmitInfo) {
	_tmp_waitInfo := *waitInfo
	atEnd := ctx.Begin("Device_Submit")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke9(libcall.t_Device_Submit, 7, uintptr(dev), uintptr(cmd), uintptr(priority), sliceToUintptr(info), uintptr(len(info)), uintptr(waitStage), uintptr(unsafe.Pointer(&_tmp_waitInfo)), 0, 0)
	handleError(ctx, rc)
	*waitInfo = _tmp_waitInfo
}
func call_Disposable_Dispose(disp hDisposable) {
	_ = dldyn.Invoke(libcall.t_Disposable_Dispose, 1, uintptr(disp), 0, 0)
}
func call_Exception_GetError(ex hException, msg []byte, msgLen *int32) {
	_tmp_msgLen := *msgLen
	_ = dldyn.Invoke6(libcall.t_Exception_GetError, 4, uintptr(ex), byteArrayToUintptr(msg), uintptr(len(msg)), uintptr(unsafe.Pointer(&_tmp_msgLen)), 0, 0)
	*msgLen = _tmp_msgLen
}
func call_GraphicsPipeline_AddAlphaBlend(ctx APIContext, pl hGraphicsPipeline) {
	atEnd := ctx.Begin("GraphicsPipeline_AddAlphaBlend")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_AddAlphaBlend, 1, uintptr(pl), 0, 0)
	handleError(ctx, rc)
}
func call_GraphicsPipeline_AddDepth(ctx APIContext, pl hGraphicsPipeline, write bool, check bool) {
	atEnd := ctx.Begin("GraphicsPipeline_AddDepth")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_AddDepth, 3, uintptr(pl), boolToUintptr(write), boolToUintptr(check))
	handleError(ctx, rc)
}
func call_GraphicsPipeline_AddVertexBinding(ctx APIContext, pl hGraphicsPipeline, stride uint32, rate VertexInputRate) {
	atEnd := ctx.Begin("GraphicsPipeline_AddVertexBinding")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_AddVertexBinding, 3, uintptr(pl), uintptr(stride), uintptr(rate))
	handleError(ctx, rc)
}
func call_GraphicsPipeline_AddVertexFormat(ctx APIContext, pl hGraphicsPipeline, format Format, offset uint32) {
	atEnd := ctx.Begin("GraphicsPipeline_AddVertexFormat")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_AddVertexFormat, 3, uintptr(pl), uintptr(format), uintptr(offset))
	handleError(ctx, rc)
}
func call_GraphicsPipeline_Create(ctx APIContext, pipeline hGraphicsPipeline, renderPass hRenderPass) {
	atEnd := ctx.Begin("GraphicsPipeline_Create")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_Create, 2, uintptr(pipeline), uintptr(renderPass), 0)
	handleError(ctx, rc)
}
func call_GraphicsPipeline_SetTopology(ctx APIContext, pl hGraphicsPipeline, topology PrimitiveTopology) {
	atEnd := ctx.Begin("GraphicsPipeline_SetTopology")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_GraphicsPipeline_SetTopology, 2, uintptr(pl), uintptr(topology), 0)
	handleError(ctx, rc)
}
func call_ImageLoader_Describe(ctx APIContext, loader hImageLoader, kind []byte, desc *ImageDescription, content []uint8) {
	_tmp_desc := *desc
	atEnd := ctx.Begin("ImageLoader_Describe")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_ImageLoader_Describe, 6, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(&_tmp_desc)), sliceToUintptr(content), uintptr(len(content)))
	handleError(ctx, rc)
	*desc = _tmp_desc
}
func call_ImageLoader_Load(ctx APIContext, loader hImageLoader, kind []byte, content []uint8, buf hBuffer) {
	atEnd := ctx.Begin("ImageLoader_Load")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_ImageLoader_Load, 6, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), sliceToUintptr(content), uintptr(len(content)), uintptr(buf))
	handleError(ctx, rc)
}
func call_ImageLoader_Save(ctx APIContext, loader hImageLoader, kind []byte, desc *ImageDescription, buf hBuffer, content []uint8, reqSize *uint64) {
	_tmp_desc := *desc
	_tmp_reqSize := *reqSize
	atEnd := ctx.Begin("ImageLoader_Save")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke9(libcall.t_ImageLoader_Save, 8, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(&_tmp_desc)), uintptr(buf), sliceToUintptr(content), uintptr(len(content)), uintptr(unsafe.Pointer(&_tmp_reqSize)), 0)
	handleError(ctx, rc)
	*desc = _tmp_desc
	*reqSize = _tmp_reqSize
}
func call_ImageLoader_Supported(ctx APIContext, loader hImageLoader, kind []byte, read *bool, write *bool) {
	_tmp_read := *read
	_tmp_write := *write
	atEnd := ctx.Begin("ImageLoader_Supported")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_ImageLoader_Supported, 5, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(&_tmp_read)), uintptr(unsafe.Pointer(&_tmp_write)), 0)
	handleError(ctx, rc)
	*read = _tmp_read
	*write = _tmp_write
}
func call_Image_NewView(ctx APIContext, image hImage, imRange *ImageRange, imageView *hImageView, cube bool) {
	_tmp_imRange := *imRange
	_tmp_imageView := *imageView
	atEnd := ctx.Begin("Image_NewView")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Image_NewView, 4, uintptr(image), uintptr(unsafe.Pointer(&_tmp_imRange)), uintptr(unsafe.Pointer(&_tmp_imageView)), boolToUintptr(cube), 0, 0)
	handleError(ctx, rc)
	*imRange = _tmp_imRange
	*imageView = _tmp_imageView
}
func call_Instance_GetPhysicalDevice(ctx APIContext, instance hInstance, index int32, info *DeviceInfo) {
	_tmp_info := *info
	atEnd := ctx.Begin("Instance_GetPhysicalDevice")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Instance_GetPhysicalDevice, 3, uintptr(instance), uintptr(index), uintptr(unsafe.Pointer(&_tmp_info)))
	handleError(ctx, rc)
	*info = _tmp_info
}
func call_Instance_NewDevice(ctx APIContext, instance hInstance, index int32, pd *hDevice) {
	_tmp_pd := *pd
	atEnd := ctx.Begin("Instance_NewDevice")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Instance_NewDevice, 3, uintptr(instance), uintptr(index), uintptr(unsafe.Pointer(&_tmp_pd)))
	handleError(ctx, rc)
	*pd = _tmp_pd
}
func call_MemoryBlock_Allocate(ctx APIContext, memBlock hMemoryBlock) {
	atEnd := ctx.Begin("MemoryBlock_Allocate")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_MemoryBlock_Allocate, 1, uintptr(memBlock), 0, 0)
	handleError(ctx, rc)
}
func call_MemoryBlock_Reserve(ctx APIContext, memBlock hMemoryBlock, memObject hMemoryObject, suitable *bool) {
	_tmp_suitable := *suitable
	atEnd := ctx.Begin("MemoryBlock_Reserve")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_MemoryBlock_Reserve, 3, uintptr(memBlock), uintptr(memObject), uintptr(unsafe.Pointer(&_tmp_suitable)))
	handleError(ctx, rc)
	*suitable = _tmp_suitable
}
func call_NewApplication(ctx APIContext, name []byte, app *hApplication) {
	_tmp_app := *app
	atEnd := ctx.Begin("NewApplication")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_NewApplication, 3, byteArrayToUintptr(name), uintptr(len(name)), uintptr(unsafe.Pointer(&_tmp_app)))
	handleError(ctx, rc)
	*app = _tmp_app
}
func call_NewDesktop(ctx APIContext, app hApplication, imageUsage ImageUsageFlags, desktop *hDesktop) {
	_tmp_desktop := *desktop
	atEnd := ctx.Begin("NewDesktop")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_NewDesktop, 3, uintptr(app), uintptr(imageUsage), uintptr(unsafe.Pointer(&_tmp_desktop)))
	handleError(ctx, rc)
	*desktop = _tmp_desktop
}
func call_NewImageLoader(ctx APIContext, loader *hImageLoader) {
	_tmp_loader := *loader
	atEnd := ctx.Begin("NewImageLoader")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_NewImageLoader, 1, uintptr(unsafe.Pointer(&_tmp_loader)), 0, 0)
	handleError(ctx, rc)
	*loader = _tmp_loader
}
func call_NewRenderPass(ctx APIContext, dev hDevice, rp *hRenderPass, depthAttachment bool, attachments []AttachmentInfo) {
	_tmp_rp := *rp
	atEnd := ctx.Begin("NewRenderPass")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_NewRenderPass, 5, uintptr(dev), uintptr(unsafe.Pointer(&_tmp_rp)), boolToUintptr(depthAttachment), sliceToUintptr(attachments), uintptr(len(attachments)), 0)
	handleError(ctx, rc)
	*rp = _tmp_rp
}
func call_Pipeline_AddDescriptorLayout(ctx APIContext, pl hPipeline, dsLayout hDescriptorLayout) {
	atEnd := ctx.Begin("Pipeline_AddDescriptorLayout")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Pipeline_AddDescriptorLayout, 2, uintptr(pl), uintptr(dsLayout), 0)
	handleError(ctx, rc)
}
func call_Pipeline_AddShader(ctx APIContext, pl hPipeline, stage ShaderStageFlags, code []uint8) {
	atEnd := ctx.Begin("Pipeline_AddShader")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Pipeline_AddShader, 4, uintptr(pl), uintptr(stage), sliceToUintptr(code), uintptr(len(code)), 0, 0)
	handleError(ctx, rc)
}
func call_QueryPool_Get(ctx APIContext, qp hQueryPool, values []uint64, timestampPeriod *float32) {
	_tmp_timestampPeriod := *timestampPeriod
	atEnd := ctx.Begin("QueryPool_Get")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_QueryPool_Get, 4, uintptr(qp), sliceToUintptr(values), uintptr(len(values)), uintptr(unsafe.Pointer(&_tmp_timestampPeriod)), 0, 0)
	handleError(ctx, rc)
	*timestampPeriod = _tmp_timestampPeriod
}
func call_RenderPass_NewFrameBuffer(ctx APIContext, rp hRenderPass, attachments []hImageView, fb *hFramebuffer) {
	_tmp_fb := *fb
	atEnd := ctx.Begin("RenderPass_NewFrameBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_RenderPass_NewFrameBuffer, 4, uintptr(rp), sliceToUintptr(attachments), uintptr(len(attachments)), uintptr(unsafe.Pointer(&_tmp_fb)), 0, 0)
	handleError(ctx, rc)
	*fb = _tmp_fb
}
func call_RenderPass_NewNullFrameBuffer(ctx APIContext, rp hRenderPass, width uint32, height uint32, fb *hFramebuffer) {
	_tmp_fb := *fb
	atEnd := ctx.Begin("RenderPass_NewNullFrameBuffer")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_RenderPass_NewNullFrameBuffer, 4, uintptr(rp), uintptr(width), uintptr(height), uintptr(unsafe.Pointer(&_tmp_fb)), 0, 0)
	handleError(ctx, rc)
	*fb = _tmp_fb
}
func call_Window_GetNextFrame(ctx APIContext, win hWindow, image *hImage, submitInfo *hSubmitInfo, viewIndex *int32) {
	_tmp_image := *image
	_tmp_submitInfo := *submitInfo
	_tmp_viewIndex := *viewIndex
	atEnd := ctx.Begin("Window_GetNextFrame")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Window_GetNextFrame, 4, uintptr(win), uintptr(unsafe.Pointer(&_tmp_image)), uintptr(unsafe.Pointer(&_tmp_submitInfo)), uintptr(unsafe.Pointer(&_tmp_viewIndex)), 0, 0)
	handleError(ctx, rc)
	*image = _tmp_image
	*submitInfo = _tmp_submitInfo
	*viewIndex = _tmp_viewIndex
}
func call_Window_GetPos(ctx APIContext, win hWindow, pos *WindowPos) {
	_tmp_pos := *pos
	atEnd := ctx.Begin("Window_GetPos")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Window_GetPos, 2, uintptr(win), uintptr(unsafe.Pointer(&_tmp_pos)), 0)
	handleError(ctx, rc)
	*pos = _tmp_pos
}
func call_Window_PrepareSwapchain(ctx APIContext, win hWindow, dev hDevice, imageDesc *ImageDescription, imageCount *int32) {
	_tmp_imageDesc := *imageDesc
	_tmp_imageCount := *imageCount
	atEnd := ctx.Begin("Window_PrepareSwapchain")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke6(libcall.t_Window_PrepareSwapchain, 4, uintptr(win), uintptr(dev), uintptr(unsafe.Pointer(&_tmp_imageDesc)), uintptr(unsafe.Pointer(&_tmp_imageCount)), 0, 0)
	handleError(ctx, rc)
	*imageDesc = _tmp_imageDesc
	*imageCount = _tmp_imageCount
}
func call_Window_SetPos(ctx APIContext, win hWindow, pos *WindowPos) {
	_tmp_pos := *pos
	atEnd := ctx.Begin("Window_SetPos")
	if atEnd != nil {
		defer atEnd()
	}
	rc := dldyn.Invoke(libcall.t_Window_SetPos, 2, uintptr(win), uintptr(unsafe.Pointer(&_tmp_pos)), 0)
	handleError(ctx, rc)
	*pos = _tmp_pos
}
