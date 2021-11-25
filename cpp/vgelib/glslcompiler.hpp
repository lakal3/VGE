#pragma once

namespace vge {
    struct CompilePass;
    class GlslCompiler: public Disposable {
    public:
        GlslCompiler(Device* dev) :_dev(dev) {
        };
        virtual void Dispose() override;
        void Compile(vk::ShaderStageFlags stage, uint8_t* src, size_t src_len, void*& instance);
        void GetOutput(void* instance, void*& msg, uint64_t& msg_len, uint32_t& result);
        void GetSpirv(void* instance, void*& spirv, uint64_t& spirv_len);
        void Free(void* instance);
    private:
        void getErrors(CompilePass *cp);
        void getLinkErrors(CompilePass* cp);
        void fillLimits(CompilePass* cp);
        Device* _dev;
    };

}