package vk
// Autogenerated file - do not edit!


import (
	"syscall"
	"unsafe"
)

var libcall struct {
    h_lib syscall.Handle

    t_AddDynamicDescriptors uintptr
    t_AddValidation uintptr
    t_AddValidationException uintptr
    t_Application_Init uintptr
    t_Buffer_GetPtr uintptr
    t_Buffer_NewView uintptr
    t_Command_Begin uintptr
    t_Command_BeginRenderPass uintptr
    t_Command_ClearImage uintptr
    t_Command_Compute uintptr
    t_Command_CopyBuffer uintptr
    t_Command_CopyBufferToImage uintptr
    t_Command_CopyImageToBuffer uintptr
    t_Command_Draw uintptr
    t_Command_EndRenderPass uintptr
    t_Command_SetLayout uintptr
    t_Command_Wait uintptr
    t_Command_WriteTimer uintptr
    t_ComputePipeline_Create uintptr
    t_DebugPoint uintptr
    t_DescriptorLayout_NewPool uintptr
    t_DescriptorPool_Alloc uintptr
    t_DescriptorSet_WriteBuffer uintptr
    t_DescriptorSet_WriteBufferView uintptr
    t_DescriptorSet_WriteImage uintptr
    t_Desktop_CreateWindow uintptr
    t_Desktop_GetKeyName uintptr
    t_Desktop_GetMonitor uintptr
    t_Desktop_PullEvent uintptr
    t_Device_NewBuffer uintptr
    t_Device_NewCommand uintptr
    t_Device_NewComputePipeline uintptr
    t_Device_NewDescriptorLayout uintptr
    t_Device_NewGraphicsPipeline uintptr
    t_Device_NewImage uintptr
    t_Device_NewMemoryBlock uintptr
    t_Device_NewSampler uintptr
    t_Device_NewTimestampQuery uintptr
    t_Device_Submit uintptr
    t_Disposable_Dispose uintptr
    t_Exception_GetError uintptr
    t_GraphicsPipeline_AddAlphaBlend uintptr
    t_GraphicsPipeline_AddDepth uintptr
    t_GraphicsPipeline_AddVertexBinding uintptr
    t_GraphicsPipeline_AddVertexFormat uintptr
    t_GraphicsPipeline_Create uintptr
    t_GraphicsPipeline_SetTopology uintptr
    t_ImageLoader_Describe uintptr
    t_ImageLoader_Load uintptr
    t_ImageLoader_Save uintptr
    t_ImageLoader_Supported uintptr
    t_Image_NewView uintptr
    t_Instance_GetPhysicalDevice uintptr
    t_Instance_NewDevice uintptr
    t_MemoryBlock_Allocate uintptr
    t_MemoryBlock_Reserve uintptr
    t_NewApplication uintptr
    t_NewDesktop uintptr
    t_NewImageLoader uintptr
    t_NewRenderPass uintptr
    t_Pipeline_AddDescriptorLayout uintptr
    t_Pipeline_AddShader uintptr
    t_QueryPool_Get uintptr
    t_RenderPass_NewFrameBuffer uintptr
    t_RenderPass_NewNullFrameBuffer uintptr
    t_Window_GetNextFrame uintptr
    t_Window_GetPos uintptr
    t_Window_PrepareSwapchain uintptr
    t_Window_SetPos uintptr
}

func loadLib() (err error) {	
	if libcall.h_lib != 0 {
		return nil
	}
	libcall.h_lib, err = syscall.LoadLibrary(GetDllPath())
	if err != nil {
		return err
	}

    libcall.t_AddDynamicDescriptors , err = syscall.GetProcAddress(libcall.h_lib, "AddDynamicDescriptors")
    if err != nil {
        return err
    }
    libcall.t_AddValidation , err = syscall.GetProcAddress(libcall.h_lib, "AddValidation")
    if err != nil {
        return err
    }
    libcall.t_AddValidationException , err = syscall.GetProcAddress(libcall.h_lib, "AddValidationException")
    if err != nil {
        return err
    }
    libcall.t_Application_Init , err = syscall.GetProcAddress(libcall.h_lib, "Application_Init")
    if err != nil {
        return err
    }
    libcall.t_Buffer_GetPtr , err = syscall.GetProcAddress(libcall.h_lib, "Buffer_GetPtr")
    if err != nil {
        return err
    }
    libcall.t_Buffer_NewView , err = syscall.GetProcAddress(libcall.h_lib, "Buffer_NewView")
    if err != nil {
        return err
    }
    libcall.t_Command_Begin , err = syscall.GetProcAddress(libcall.h_lib, "Command_Begin")
    if err != nil {
        return err
    }
    libcall.t_Command_BeginRenderPass , err = syscall.GetProcAddress(libcall.h_lib, "Command_BeginRenderPass")
    if err != nil {
        return err
    }
    libcall.t_Command_ClearImage , err = syscall.GetProcAddress(libcall.h_lib, "Command_ClearImage")
    if err != nil {
        return err
    }
    libcall.t_Command_Compute , err = syscall.GetProcAddress(libcall.h_lib, "Command_Compute")
    if err != nil {
        return err
    }
    libcall.t_Command_CopyBuffer , err = syscall.GetProcAddress(libcall.h_lib, "Command_CopyBuffer")
    if err != nil {
        return err
    }
    libcall.t_Command_CopyBufferToImage , err = syscall.GetProcAddress(libcall.h_lib, "Command_CopyBufferToImage")
    if err != nil {
        return err
    }
    libcall.t_Command_CopyImageToBuffer , err = syscall.GetProcAddress(libcall.h_lib, "Command_CopyImageToBuffer")
    if err != nil {
        return err
    }
    libcall.t_Command_Draw , err = syscall.GetProcAddress(libcall.h_lib, "Command_Draw")
    if err != nil {
        return err
    }
    libcall.t_Command_EndRenderPass , err = syscall.GetProcAddress(libcall.h_lib, "Command_EndRenderPass")
    if err != nil {
        return err
    }
    libcall.t_Command_SetLayout , err = syscall.GetProcAddress(libcall.h_lib, "Command_SetLayout")
    if err != nil {
        return err
    }
    libcall.t_Command_Wait , err = syscall.GetProcAddress(libcall.h_lib, "Command_Wait")
    if err != nil {
        return err
    }
    libcall.t_Command_WriteTimer , err = syscall.GetProcAddress(libcall.h_lib, "Command_WriteTimer")
    if err != nil {
        return err
    }
    libcall.t_ComputePipeline_Create , err = syscall.GetProcAddress(libcall.h_lib, "ComputePipeline_Create")
    if err != nil {
        return err
    }
    libcall.t_DebugPoint , err = syscall.GetProcAddress(libcall.h_lib, "DebugPoint")
    if err != nil {
        return err
    }
    libcall.t_DescriptorLayout_NewPool , err = syscall.GetProcAddress(libcall.h_lib, "DescriptorLayout_NewPool")
    if err != nil {
        return err
    }
    libcall.t_DescriptorPool_Alloc , err = syscall.GetProcAddress(libcall.h_lib, "DescriptorPool_Alloc")
    if err != nil {
        return err
    }
    libcall.t_DescriptorSet_WriteBuffer , err = syscall.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteBuffer")
    if err != nil {
        return err
    }
    libcall.t_DescriptorSet_WriteBufferView , err = syscall.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteBufferView")
    if err != nil {
        return err
    }
    libcall.t_DescriptorSet_WriteImage , err = syscall.GetProcAddress(libcall.h_lib, "DescriptorSet_WriteImage")
    if err != nil {
        return err
    }
    libcall.t_Desktop_CreateWindow , err = syscall.GetProcAddress(libcall.h_lib, "Desktop_CreateWindow")
    if err != nil {
        return err
    }
    libcall.t_Desktop_GetKeyName , err = syscall.GetProcAddress(libcall.h_lib, "Desktop_GetKeyName")
    if err != nil {
        return err
    }
    libcall.t_Desktop_GetMonitor , err = syscall.GetProcAddress(libcall.h_lib, "Desktop_GetMonitor")
    if err != nil {
        return err
    }
    libcall.t_Desktop_PullEvent , err = syscall.GetProcAddress(libcall.h_lib, "Desktop_PullEvent")
    if err != nil {
        return err
    }
    libcall.t_Device_NewBuffer , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewBuffer")
    if err != nil {
        return err
    }
    libcall.t_Device_NewCommand , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewCommand")
    if err != nil {
        return err
    }
    libcall.t_Device_NewComputePipeline , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewComputePipeline")
    if err != nil {
        return err
    }
    libcall.t_Device_NewDescriptorLayout , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewDescriptorLayout")
    if err != nil {
        return err
    }
    libcall.t_Device_NewGraphicsPipeline , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewGraphicsPipeline")
    if err != nil {
        return err
    }
    libcall.t_Device_NewImage , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewImage")
    if err != nil {
        return err
    }
    libcall.t_Device_NewMemoryBlock , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewMemoryBlock")
    if err != nil {
        return err
    }
    libcall.t_Device_NewSampler , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewSampler")
    if err != nil {
        return err
    }
    libcall.t_Device_NewTimestampQuery , err = syscall.GetProcAddress(libcall.h_lib, "Device_NewTimestampQuery")
    if err != nil {
        return err
    }
    libcall.t_Device_Submit , err = syscall.GetProcAddress(libcall.h_lib, "Device_Submit")
    if err != nil {
        return err
    }
    libcall.t_Disposable_Dispose , err = syscall.GetProcAddress(libcall.h_lib, "Disposable_Dispose")
    if err != nil {
        return err
    }
    libcall.t_Exception_GetError , err = syscall.GetProcAddress(libcall.h_lib, "Exception_GetError")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_AddAlphaBlend , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddAlphaBlend")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_AddDepth , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddDepth")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_AddVertexBinding , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddVertexBinding")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_AddVertexFormat , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_AddVertexFormat")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_Create , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_Create")
    if err != nil {
        return err
    }
    libcall.t_GraphicsPipeline_SetTopology , err = syscall.GetProcAddress(libcall.h_lib, "GraphicsPipeline_SetTopology")
    if err != nil {
        return err
    }
    libcall.t_ImageLoader_Describe , err = syscall.GetProcAddress(libcall.h_lib, "ImageLoader_Describe")
    if err != nil {
        return err
    }
    libcall.t_ImageLoader_Load , err = syscall.GetProcAddress(libcall.h_lib, "ImageLoader_Load")
    if err != nil {
        return err
    }
    libcall.t_ImageLoader_Save , err = syscall.GetProcAddress(libcall.h_lib, "ImageLoader_Save")
    if err != nil {
        return err
    }
    libcall.t_ImageLoader_Supported , err = syscall.GetProcAddress(libcall.h_lib, "ImageLoader_Supported")
    if err != nil {
        return err
    }
    libcall.t_Image_NewView , err = syscall.GetProcAddress(libcall.h_lib, "Image_NewView")
    if err != nil {
        return err
    }
    libcall.t_Instance_GetPhysicalDevice , err = syscall.GetProcAddress(libcall.h_lib, "Instance_GetPhysicalDevice")
    if err != nil {
        return err
    }
    libcall.t_Instance_NewDevice , err = syscall.GetProcAddress(libcall.h_lib, "Instance_NewDevice")
    if err != nil {
        return err
    }
    libcall.t_MemoryBlock_Allocate , err = syscall.GetProcAddress(libcall.h_lib, "MemoryBlock_Allocate")
    if err != nil {
        return err
    }
    libcall.t_MemoryBlock_Reserve , err = syscall.GetProcAddress(libcall.h_lib, "MemoryBlock_Reserve")
    if err != nil {
        return err
    }
    libcall.t_NewApplication , err = syscall.GetProcAddress(libcall.h_lib, "NewApplication")
    if err != nil {
        return err
    }
    libcall.t_NewDesktop , err = syscall.GetProcAddress(libcall.h_lib, "NewDesktop")
    if err != nil {
        return err
    }
    libcall.t_NewImageLoader , err = syscall.GetProcAddress(libcall.h_lib, "NewImageLoader")
    if err != nil {
        return err
    }
    libcall.t_NewRenderPass , err = syscall.GetProcAddress(libcall.h_lib, "NewRenderPass")
    if err != nil {
        return err
    }
    libcall.t_Pipeline_AddDescriptorLayout , err = syscall.GetProcAddress(libcall.h_lib, "Pipeline_AddDescriptorLayout")
    if err != nil {
        return err
    }
    libcall.t_Pipeline_AddShader , err = syscall.GetProcAddress(libcall.h_lib, "Pipeline_AddShader")
    if err != nil {
        return err
    }
    libcall.t_QueryPool_Get , err = syscall.GetProcAddress(libcall.h_lib, "QueryPool_Get")
    if err != nil {
        return err
    }
    libcall.t_RenderPass_NewFrameBuffer , err = syscall.GetProcAddress(libcall.h_lib, "RenderPass_NewFrameBuffer")
    if err != nil {
        return err
    }
    libcall.t_RenderPass_NewNullFrameBuffer , err = syscall.GetProcAddress(libcall.h_lib, "RenderPass_NewNullFrameBuffer")
    if err != nil {
        return err
    }
    libcall.t_Window_GetNextFrame , err = syscall.GetProcAddress(libcall.h_lib, "Window_GetNextFrame")
    if err != nil {
        return err
    }
    libcall.t_Window_GetPos , err = syscall.GetProcAddress(libcall.h_lib, "Window_GetPos")
    if err != nil {
        return err
    }
    libcall.t_Window_PrepareSwapchain , err = syscall.GetProcAddress(libcall.h_lib, "Window_PrepareSwapchain")
    if err != nil {
        return err
    }
    libcall.t_Window_SetPos , err = syscall.GetProcAddress(libcall.h_lib, "Window_SetPos")
    if err != nil {
        return err
    }
    return nil
}



func call_AddDynamicDescriptors(ctx APIContext,app hApplication) {
    atEnd := ctx.Begin("AddDynamicDescriptors")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_AddDynamicDescriptors,1, uintptr(app),0,0)
    handleError(ctx, rc)
}
func call_AddValidation(ctx APIContext,app hApplication) {
    atEnd := ctx.Begin("AddValidation")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_AddValidation,1, uintptr(app),0,0)
    handleError(ctx, rc)
}
func call_AddValidationException(ctx APIContext,msgId int32) {
    atEnd := ctx.Begin("AddValidationException")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_AddValidationException,1, uintptr(msgId),0,0)
    handleError(ctx, rc)
}
func call_Application_Init(ctx APIContext,app hApplication, inst *hInstance) {
    atEnd := ctx.Begin("Application_Init")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Application_Init,2, uintptr(app), uintptr(unsafe.Pointer(inst)),0)
    handleError(ctx, rc)
}
func call_Buffer_GetPtr(ctx APIContext,buffer hBuffer, ptr *uintptr) {
    atEnd := ctx.Begin("Buffer_GetPtr")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Buffer_GetPtr,2, uintptr(buffer), uintptr(unsafe.Pointer(ptr)),0)
    handleError(ctx, rc)
}
func call_Buffer_NewView(ctx APIContext,buffer hBuffer, format Format, offset uint64, size uint64, view *hBufferView) {
    atEnd := ctx.Begin("Buffer_NewView")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Buffer_NewView,5, uintptr(buffer), uintptr(format), uintptr(offset), uintptr(size), uintptr(unsafe.Pointer(view)),0)
    handleError(ctx, rc)
}
func call_Command_Begin(ctx APIContext,cmd hCommand) {
    atEnd := ctx.Begin("Command_Begin")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_Begin,1, uintptr(cmd),0,0)
    handleError(ctx, rc)
}
func call_Command_BeginRenderPass(ctx APIContext,cmd hCommand, rp hRenderPass, fb hFramebuffer) {
    atEnd := ctx.Begin("Command_BeginRenderPass")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_BeginRenderPass,3, uintptr(cmd), uintptr(rp), uintptr(fb))
    handleError(ctx, rc)
}
func call_Command_ClearImage(ctx APIContext,cmd hCommand, dst hImage, imRange *ImageRange, layout ImageLayout, color float32, alpha float32) {
    atEnd := ctx.Begin("Command_ClearImage")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Command_ClearImage,6, uintptr(cmd), uintptr(dst), uintptr(unsafe.Pointer(imRange)), uintptr(layout), uintptr(color), uintptr(alpha))
    handleError(ctx, rc)
}
func call_Command_Compute(ctx APIContext,hCmd hCommand, hPl hComputePipeline, x uint32, y uint32, z uint32, descriptors []hDescriptorSet) {
    atEnd := ctx.Begin("Command_Compute")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall9(libcall.t_Command_Compute,7, uintptr(hCmd), uintptr(hPl), uintptr(x), uintptr(y), uintptr(z), sliceToUintptr(descriptors), uintptr(len(descriptors)),0,0)
    handleError(ctx, rc)
}
func call_Command_CopyBuffer(ctx APIContext,cmd hCommand, src hBuffer, dst hBuffer) {
    atEnd := ctx.Begin("Command_CopyBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_CopyBuffer,3, uintptr(cmd), uintptr(src), uintptr(dst))
    handleError(ctx, rc)
}
func call_Command_CopyBufferToImage(ctx APIContext,cmd hCommand, src hBuffer, dst hImage, imRange *ImageRange, offset uint64) {
    atEnd := ctx.Begin("Command_CopyBufferToImage")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Command_CopyBufferToImage,5, uintptr(cmd), uintptr(src), uintptr(dst), uintptr(unsafe.Pointer(imRange)), uintptr(offset),0)
    handleError(ctx, rc)
}
func call_Command_CopyImageToBuffer(ctx APIContext,cmd hCommand, src hImage, dst hBuffer, imRange *ImageRange, offset uint64) {
    atEnd := ctx.Begin("Command_CopyImageToBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Command_CopyImageToBuffer,5, uintptr(cmd), uintptr(src), uintptr(dst), uintptr(unsafe.Pointer(imRange)), uintptr(offset),0)
    handleError(ctx, rc)
}
func call_Command_Draw(ctx APIContext,cmd hCommand, draws []DrawItem) {
    atEnd := ctx.Begin("Command_Draw")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_Draw,3, uintptr(cmd), sliceToUintptr(draws), uintptr(len(draws)))
    handleError(ctx, rc)
}
func call_Command_EndRenderPass(ctx APIContext,cmd hCommand) {
    atEnd := ctx.Begin("Command_EndRenderPass")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_EndRenderPass,1, uintptr(cmd),0,0)
    handleError(ctx, rc)
}
func call_Command_SetLayout(ctx APIContext,cmd hCommand, image hImage, imRange *ImageRange, newLayout ImageLayout) {
    atEnd := ctx.Begin("Command_SetLayout")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Command_SetLayout,4, uintptr(cmd), uintptr(image), uintptr(unsafe.Pointer(imRange)), uintptr(newLayout),0,0)
    handleError(ctx, rc)
}
func call_Command_Wait(ctx APIContext,cmd hCommand) {
    atEnd := ctx.Begin("Command_Wait")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Command_Wait,1, uintptr(cmd),0,0)
    handleError(ctx, rc)
}
func call_Command_WriteTimer(ctx APIContext,cmd hCommand, qp hQueryPool, stages PipelineStageFlags, timerIndex uint32) {
    atEnd := ctx.Begin("Command_WriteTimer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Command_WriteTimer,4, uintptr(cmd), uintptr(qp), uintptr(stages), uintptr(timerIndex),0,0)
    handleError(ctx, rc)
}
func call_ComputePipeline_Create(ctx APIContext,cp hComputePipeline) {
    atEnd := ctx.Begin("ComputePipeline_Create")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_ComputePipeline_Create,1, uintptr(cp),0,0)
    handleError(ctx, rc)
}
func call_DebugPoint(point []byte) {
    _, _, _ = syscall.Syscall(libcall.t_DebugPoint,2, byteArrayToUintptr(point), uintptr(len(point)),0)
}
func call_DescriptorLayout_NewPool(ctx APIContext,layout hDescriptorLayout, size uint32, pool *hDescriptorPool) {
    atEnd := ctx.Begin("DescriptorLayout_NewPool")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_DescriptorLayout_NewPool,3, uintptr(layout), uintptr(size), uintptr(unsafe.Pointer(pool)))
    handleError(ctx, rc)
}
func call_DescriptorPool_Alloc(ctx APIContext,pool hDescriptorPool, ds *hDescriptorSet) {
    atEnd := ctx.Begin("DescriptorPool_Alloc")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_DescriptorPool_Alloc,2, uintptr(pool), uintptr(unsafe.Pointer(ds)),0)
    handleError(ctx, rc)
}
func call_DescriptorSet_WriteBuffer(ctx APIContext,ds hDescriptorSet, binding uint32, at uint32, buffer hBuffer, from uint64, size uint64) {
    atEnd := ctx.Begin("DescriptorSet_WriteBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_DescriptorSet_WriteBuffer,6, uintptr(ds), uintptr(binding), uintptr(at), uintptr(buffer), uintptr(from), uintptr(size))
    handleError(ctx, rc)
}
func call_DescriptorSet_WriteBufferView(ctx APIContext,ds hDescriptorSet, binding uint32, at uint32, bufferView hBufferView) {
    atEnd := ctx.Begin("DescriptorSet_WriteBufferView")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_DescriptorSet_WriteBufferView,4, uintptr(ds), uintptr(binding), uintptr(at), uintptr(bufferView),0,0)
    handleError(ctx, rc)
}
func call_DescriptorSet_WriteImage(ctx APIContext,ds hDescriptorSet, binding uint32, at uint32, view hImageView, sampler hSampler) {
    atEnd := ctx.Begin("DescriptorSet_WriteImage")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_DescriptorSet_WriteImage,5, uintptr(ds), uintptr(binding), uintptr(at), uintptr(view), uintptr(sampler),0)
    handleError(ctx, rc)
}
func call_Desktop_CreateWindow(ctx APIContext,desktop hDesktop, title []byte, pos *WindowPos, win *hWindow) {
    atEnd := ctx.Begin("Desktop_CreateWindow")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Desktop_CreateWindow,5, uintptr(desktop), byteArrayToUintptr(title), uintptr(len(title)), uintptr(unsafe.Pointer(pos)), uintptr(unsafe.Pointer(win)),0)
    handleError(ctx, rc)
}
func call_Desktop_GetKeyName(ctx APIContext,desktop hDesktop, keyCode uint32, name []uint8, strLen *uint32) {
    atEnd := ctx.Begin("Desktop_GetKeyName")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Desktop_GetKeyName,5, uintptr(desktop), uintptr(keyCode), sliceToUintptr(name), uintptr(len(name)), uintptr(unsafe.Pointer(strLen)),0)
    handleError(ctx, rc)
}
func call_Desktop_GetMonitor(ctx APIContext,desktop hDesktop, monitor uint32, info *WindowPos) {
    atEnd := ctx.Begin("Desktop_GetMonitor")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Desktop_GetMonitor,3, uintptr(desktop), uintptr(monitor), uintptr(unsafe.Pointer(info)))
    handleError(ctx, rc)
}
func call_Desktop_PullEvent(ctx APIContext,desktop hDesktop, ev *RawEvent) {
    atEnd := ctx.Begin("Desktop_PullEvent")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Desktop_PullEvent,2, uintptr(desktop), uintptr(unsafe.Pointer(ev)),0)
    handleError(ctx, rc)
}
func call_Device_NewBuffer(ctx APIContext,dev hDevice, size uint64, hostMemory bool, usage BufferUsageFlags, buffer *hBuffer) {
    atEnd := ctx.Begin("Device_NewBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Device_NewBuffer,5, uintptr(dev), uintptr(size), boolToUintptr(hostMemory), uintptr(usage), uintptr(unsafe.Pointer(buffer)),0)
    handleError(ctx, rc)
}
func call_Device_NewCommand(ctx APIContext,dev hDevice, queueType QueueFlags, once bool, command *hCommand) {
    atEnd := ctx.Begin("Device_NewCommand")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Device_NewCommand,4, uintptr(dev), uintptr(queueType), boolToUintptr(once), uintptr(unsafe.Pointer(command)),0,0)
    handleError(ctx, rc)
}
func call_Device_NewComputePipeline(ctx APIContext,dev hDevice, cp *hComputePipeline) {
    atEnd := ctx.Begin("Device_NewComputePipeline")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Device_NewComputePipeline,2, uintptr(dev), uintptr(unsafe.Pointer(cp)),0)
    handleError(ctx, rc)
}
func call_Device_NewDescriptorLayout(ctx APIContext,dev hDevice, descriptorType DescriptorType, stages ShaderStageFlags, element uint32, flags DescriptorBindingFlagBitsEXT, prevLayout hDescriptorLayout, dsLayout *hDescriptorLayout) {
    atEnd := ctx.Begin("Device_NewDescriptorLayout")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall9(libcall.t_Device_NewDescriptorLayout,7, uintptr(dev), uintptr(descriptorType), uintptr(stages), uintptr(element), uintptr(flags), uintptr(prevLayout), uintptr(unsafe.Pointer(dsLayout)),0,0)
    handleError(ctx, rc)
}
func call_Device_NewGraphicsPipeline(ctx APIContext,dev hDevice, gp *hGraphicsPipeline) {
    atEnd := ctx.Begin("Device_NewGraphicsPipeline")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Device_NewGraphicsPipeline,2, uintptr(dev), uintptr(unsafe.Pointer(gp)),0)
    handleError(ctx, rc)
}
func call_Device_NewImage(ctx APIContext,dev hDevice, usage ImageUsageFlags, desc *ImageDescription, image *hImage) {
    atEnd := ctx.Begin("Device_NewImage")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Device_NewImage,4, uintptr(dev), uintptr(usage), uintptr(unsafe.Pointer(desc)), uintptr(unsafe.Pointer(image)),0,0)
    handleError(ctx, rc)
}
func call_Device_NewMemoryBlock(ctx APIContext,dev hDevice, memBlock *hMemoryBlock) {
    atEnd := ctx.Begin("Device_NewMemoryBlock")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Device_NewMemoryBlock,2, uintptr(dev), uintptr(unsafe.Pointer(memBlock)),0)
    handleError(ctx, rc)
}
func call_Device_NewSampler(ctx APIContext,dev hDevice, repeatMode SamplerAddressMode, sampler *hSampler) {
    atEnd := ctx.Begin("Device_NewSampler")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Device_NewSampler,3, uintptr(dev), uintptr(repeatMode), uintptr(unsafe.Pointer(sampler)))
    handleError(ctx, rc)
}
func call_Device_NewTimestampQuery(ctx APIContext,dev hDevice, size uint32, qp *hQueryPool) {
    atEnd := ctx.Begin("Device_NewTimestampQuery")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Device_NewTimestampQuery,3, uintptr(dev), uintptr(size), uintptr(unsafe.Pointer(qp)))
    handleError(ctx, rc)
}
func call_Device_Submit(ctx APIContext,dev hDevice, cmd hCommand, priority uint32, info []hSubmitInfo, waitStage PipelineStageFlags, waitInfo *hSubmitInfo) {
    atEnd := ctx.Begin("Device_Submit")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall9(libcall.t_Device_Submit,7, uintptr(dev), uintptr(cmd), uintptr(priority), sliceToUintptr(info), uintptr(len(info)), uintptr(waitStage), uintptr(unsafe.Pointer(waitInfo)),0,0)
    handleError(ctx, rc)
}
func call_Disposable_Dispose(disp hDisposable) {
    _, _, _ = syscall.Syscall(libcall.t_Disposable_Dispose,1, uintptr(disp),0,0)
}
func call_Exception_GetError(ex hException, msg []byte, msgLen *int32) {
    _, _, _ = syscall.Syscall6(libcall.t_Exception_GetError,4, uintptr(ex), byteArrayToUintptr(msg), uintptr(len(msg)), uintptr(unsafe.Pointer(msgLen)),0,0)
}
func call_GraphicsPipeline_AddAlphaBlend(ctx APIContext,pl hGraphicsPipeline) {
    atEnd := ctx.Begin("GraphicsPipeline_AddAlphaBlend")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_AddAlphaBlend,1, uintptr(pl),0,0)
    handleError(ctx, rc)
}
func call_GraphicsPipeline_AddDepth(ctx APIContext,pl hGraphicsPipeline, write bool, check bool) {
    atEnd := ctx.Begin("GraphicsPipeline_AddDepth")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_AddDepth,3, uintptr(pl), boolToUintptr(write), boolToUintptr(check))
    handleError(ctx, rc)
}
func call_GraphicsPipeline_AddVertexBinding(ctx APIContext,pl hGraphicsPipeline, stride uint32, rate VertexInputRate) {
    atEnd := ctx.Begin("GraphicsPipeline_AddVertexBinding")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_AddVertexBinding,3, uintptr(pl), uintptr(stride), uintptr(rate))
    handleError(ctx, rc)
}
func call_GraphicsPipeline_AddVertexFormat(ctx APIContext,pl hGraphicsPipeline, format Format, offset uint32) {
    atEnd := ctx.Begin("GraphicsPipeline_AddVertexFormat")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_AddVertexFormat,3, uintptr(pl), uintptr(format), uintptr(offset))
    handleError(ctx, rc)
}
func call_GraphicsPipeline_Create(ctx APIContext,pipeline hGraphicsPipeline, renderPass hRenderPass) {
    atEnd := ctx.Begin("GraphicsPipeline_Create")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_Create,2, uintptr(pipeline), uintptr(renderPass),0)
    handleError(ctx, rc)
}
func call_GraphicsPipeline_SetTopology(ctx APIContext,pl hGraphicsPipeline, topology PrimitiveTopology) {
    atEnd := ctx.Begin("GraphicsPipeline_SetTopology")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_GraphicsPipeline_SetTopology,2, uintptr(pl), uintptr(topology),0)
    handleError(ctx, rc)
}
func call_ImageLoader_Describe(ctx APIContext,loader hImageLoader, kind []byte, desc *ImageDescription, content []uint8) {
    atEnd := ctx.Begin("ImageLoader_Describe")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_ImageLoader_Describe,6, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(desc)), sliceToUintptr(content), uintptr(len(content)))
    handleError(ctx, rc)
}
func call_ImageLoader_Load(ctx APIContext,loader hImageLoader, kind []byte, content []uint8, buf hBuffer) {
    atEnd := ctx.Begin("ImageLoader_Load")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_ImageLoader_Load,6, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), sliceToUintptr(content), uintptr(len(content)), uintptr(buf))
    handleError(ctx, rc)
}
func call_ImageLoader_Save(ctx APIContext,loader hImageLoader, kind []byte, desc *ImageDescription, buf hBuffer, content []uint8, reqSize *uint64) {
    atEnd := ctx.Begin("ImageLoader_Save")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall9(libcall.t_ImageLoader_Save,8, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(desc)), uintptr(buf), sliceToUintptr(content), uintptr(len(content)), uintptr(unsafe.Pointer(reqSize)),0)
    handleError(ctx, rc)
}
func call_ImageLoader_Supported(ctx APIContext,loader hImageLoader, kind []byte, read *bool, write *bool) {
    atEnd := ctx.Begin("ImageLoader_Supported")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_ImageLoader_Supported,5, uintptr(loader), byteArrayToUintptr(kind), uintptr(len(kind)), uintptr(unsafe.Pointer(read)), uintptr(unsafe.Pointer(write)),0)
    handleError(ctx, rc)
}
func call_Image_NewView(ctx APIContext,image hImage, imRange *ImageRange, imageView *hImageView, cube bool) {
    atEnd := ctx.Begin("Image_NewView")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Image_NewView,4, uintptr(image), uintptr(unsafe.Pointer(imRange)), uintptr(unsafe.Pointer(imageView)), boolToUintptr(cube),0,0)
    handleError(ctx, rc)
}
func call_Instance_GetPhysicalDevice(ctx APIContext,instance hInstance, index int32, info *DeviceInfo) {
    atEnd := ctx.Begin("Instance_GetPhysicalDevice")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Instance_GetPhysicalDevice,3, uintptr(instance), uintptr(index), uintptr(unsafe.Pointer(info)))
    handleError(ctx, rc)
}
func call_Instance_NewDevice(ctx APIContext,instance hInstance, index int32, pd *hDevice) {
    atEnd := ctx.Begin("Instance_NewDevice")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Instance_NewDevice,3, uintptr(instance), uintptr(index), uintptr(unsafe.Pointer(pd)))
    handleError(ctx, rc)
}
func call_MemoryBlock_Allocate(ctx APIContext,memBlock hMemoryBlock) {
    atEnd := ctx.Begin("MemoryBlock_Allocate")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_MemoryBlock_Allocate,1, uintptr(memBlock),0,0)
    handleError(ctx, rc)
}
func call_MemoryBlock_Reserve(ctx APIContext,memBlock hMemoryBlock, memObject hMemoryObject, suitable *bool) {
    atEnd := ctx.Begin("MemoryBlock_Reserve")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_MemoryBlock_Reserve,3, uintptr(memBlock), uintptr(memObject), uintptr(unsafe.Pointer(suitable)))
    handleError(ctx, rc)
}
func call_NewApplication(ctx APIContext,name []byte, app *hApplication) {
    atEnd := ctx.Begin("NewApplication")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_NewApplication,3, byteArrayToUintptr(name), uintptr(len(name)), uintptr(unsafe.Pointer(app)))
    handleError(ctx, rc)
}
func call_NewDesktop(ctx APIContext,app hApplication, imageUsage ImageUsageFlags, desktop *hDesktop) {
    atEnd := ctx.Begin("NewDesktop")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_NewDesktop,3, uintptr(app), uintptr(imageUsage), uintptr(unsafe.Pointer(desktop)))
    handleError(ctx, rc)
}
func call_NewImageLoader(ctx APIContext,loader *hImageLoader) {
    atEnd := ctx.Begin("NewImageLoader")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_NewImageLoader,1, uintptr(unsafe.Pointer(loader)),0,0)
    handleError(ctx, rc)
}
func call_NewRenderPass(ctx APIContext,dev hDevice, rp *hRenderPass, depthAttachment bool, attachments []AttachmentInfo) {
    atEnd := ctx.Begin("NewRenderPass")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_NewRenderPass,5, uintptr(dev), uintptr(unsafe.Pointer(rp)), boolToUintptr(depthAttachment), sliceToUintptr(attachments), uintptr(len(attachments)),0)
    handleError(ctx, rc)
}
func call_Pipeline_AddDescriptorLayout(ctx APIContext,pl hPipeline, dsLayout hDescriptorLayout) {
    atEnd := ctx.Begin("Pipeline_AddDescriptorLayout")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Pipeline_AddDescriptorLayout,2, uintptr(pl), uintptr(dsLayout),0)
    handleError(ctx, rc)
}
func call_Pipeline_AddShader(ctx APIContext,pl hPipeline, stage ShaderStageFlags, code []uint8) {
    atEnd := ctx.Begin("Pipeline_AddShader")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Pipeline_AddShader,4, uintptr(pl), uintptr(stage), sliceToUintptr(code), uintptr(len(code)),0,0)
    handleError(ctx, rc)
}
func call_QueryPool_Get(ctx APIContext,qp hQueryPool, values []uint64, timestampPeriod *float32) {
    atEnd := ctx.Begin("QueryPool_Get")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_QueryPool_Get,4, uintptr(qp), sliceToUintptr(values), uintptr(len(values)), uintptr(unsafe.Pointer(timestampPeriod)),0,0)
    handleError(ctx, rc)
}
func call_RenderPass_NewFrameBuffer(ctx APIContext,rp hRenderPass, attachments []hImageView, fb *hFramebuffer) {
    atEnd := ctx.Begin("RenderPass_NewFrameBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_RenderPass_NewFrameBuffer,4, uintptr(rp), sliceToUintptr(attachments), uintptr(len(attachments)), uintptr(unsafe.Pointer(fb)),0,0)
    handleError(ctx, rc)
}
func call_RenderPass_NewNullFrameBuffer(ctx APIContext,rp hRenderPass, width uint32, height uint32, fb *hFramebuffer) {
    atEnd := ctx.Begin("RenderPass_NewNullFrameBuffer")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_RenderPass_NewNullFrameBuffer,4, uintptr(rp), uintptr(width), uintptr(height), uintptr(unsafe.Pointer(fb)),0,0)
    handleError(ctx, rc)
}
func call_Window_GetNextFrame(ctx APIContext,win hWindow, image *hImage, submitInfo *hSubmitInfo, viewIndex *int32) {
    atEnd := ctx.Begin("Window_GetNextFrame")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Window_GetNextFrame,4, uintptr(win), uintptr(unsafe.Pointer(image)), uintptr(unsafe.Pointer(submitInfo)), uintptr(unsafe.Pointer(viewIndex)),0,0)
    handleError(ctx, rc)
}
func call_Window_GetPos(ctx APIContext,win hWindow, pos *WindowPos) {
    atEnd := ctx.Begin("Window_GetPos")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Window_GetPos,2, uintptr(win), uintptr(unsafe.Pointer(pos)),0)
    handleError(ctx, rc)
}
func call_Window_PrepareSwapchain(ctx APIContext,win hWindow, dev hDevice, imageDesc *ImageDescription, imageCount *int32) {
    atEnd := ctx.Begin("Window_PrepareSwapchain")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall6(libcall.t_Window_PrepareSwapchain,4, uintptr(win), uintptr(dev), uintptr(unsafe.Pointer(imageDesc)), uintptr(unsafe.Pointer(imageCount)),0,0)
    handleError(ctx, rc)
}
func call_Window_SetPos(ctx APIContext,win hWindow, pos *WindowPos) {
    atEnd := ctx.Begin("Window_SetPos")
    if atEnd != nil {
    	defer atEnd()
    }
    rc, _, _ := syscall.Syscall(libcall.t_Window_SetPos,2, uintptr(win), uintptr(unsafe.Pointer(pos)),0)
    handleError(ctx, rc)
}
