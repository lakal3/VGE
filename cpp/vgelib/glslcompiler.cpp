#include "vgelib/vgelib.hpp"

#include "glslang_c_interface.h"
#include "glslcompiler.hpp"

/*
const glslang_input_t input =
{
    .language = GLSLANG_SOURCE_GLSL,
    .stage = GLSLANG_STAGE_VERTEX,
    .client = GLSLANG_CLIENT_VULKAN,
    .client_version = GLSLANG_TARGET_VULKAN_1_1,
    .target_language = GLSLANG_TARGET_SPV,
    .target_language_version = GLSLANG_TARGET_SPV_1_3,
    .code = shaderCodeVertex,
    .default_version = 100,
    .default_profile = GLSLANG_NO_PROFILE,
    .force_default_version_and_profile = false,
    .forward_compatible = false,
    .messages = GLSLANG_MSG_DEFAULT_BIT,
};
*/

namespace vge {
    struct CompilePass {
        CompilePass() {
            input.language = GLSLANG_SOURCE_GLSL;
            input.client = GLSLANG_CLIENT_VULKAN;
            input.client_version = GLSLANG_TARGET_VULKAN_1_1;
            input.target_language = GLSLANG_TARGET_SPV;
            input.target_language_version = GLSLANG_TARGET_SPV_1_3;
            input.default_version = 100;
            input.default_profile = GLSLANG_NO_PROFILE;
            input.force_default_version_and_profile = false;
            input.forward_compatible = false;
            input.messages = GLSLANG_MSG_VULKAN_RULES_BIT;
            initResources();
            input.resource = &resources;
            result = 0;
        }

        void initResources() {
            // Resources copied from glslangValidator
            resources = { 
                /* .MaxLights = */ 32,
                    /* .MaxClipPlanes = */ 6,
                    /* .MaxTextureUnits = */ 32,
                    /* .MaxTextureCoords = */ 32,
                    /* .MaxVertexAttribs = */ 64,
                    /* .MaxVertexUniformComponents = */ 4096,
                    /* .MaxVaryingFloats = */ 64,
                    /* .MaxVertexTextureImageUnits = */ 32,
                    /* .MaxCombinedTextureImageUnits = */ 80,
                    /* .MaxTextureImageUnits = */ 32,
                    /* .MaxFragmentUniformComponents = */ 4096,
                    /* .MaxDrawBuffers = */ 32,
                    /* .MaxVertexUniformVectors = */ 128,
                    /* .MaxVaryingVectors = */ 8,
                    /* .MaxFragmentUniformVectors = */ 16,
                    /* .MaxVertexOutputVectors = */ 16,
                    /* .MaxFragmentInputVectors = */ 15,
                    /* .MinProgramTexelOffset = */ -8,
                    /* .MaxProgramTexelOffset = */ 7,
                    /* .MaxClipDistances = */ 8,
                    /* .MaxComputeWorkGroupCountX = */ 65535,
                    /* .MaxComputeWorkGroupCountY = */ 65535,
                    /* .MaxComputeWorkGroupCountZ = */ 65535,
                    /* .MaxComputeWorkGroupSizeX = */ 1024,
                    /* .MaxComputeWorkGroupSizeY = */ 1024,
                    /* .MaxComputeWorkGroupSizeZ = */ 64,
                    /* .MaxComputeUniformComponents = */ 1024,
                    /* .MaxComputeTextureImageUnits = */ 16,
                    /* .MaxComputeImageUniforms = */ 8,
                    /* .MaxComputeAtomicCounters = */ 8,
                    /* .MaxComputeAtomicCounterBuffers = */ 1,
                    /* .MaxVaryingComponents = */ 60,
                    /* .MaxVertexOutputComponents = */ 64,
                    /* .MaxGeometryInputComponents = */ 64,
                    /* .MaxGeometryOutputComponents = */ 128,
                    /* .MaxFragmentInputComponents = */ 128,
                    /* .MaxImageUnits = */ 8,
                    /* .MaxCombinedImageUnitsAndFragmentOutputs = */ 8,
                    /* .MaxCombinedShaderOutputResources = */ 8,
                    /* .MaxImageSamples = */ 0,
                    /* .MaxVertexImageUniforms = */ 0,
                    /* .MaxTessControlImageUniforms = */ 0,
                    /* .MaxTessEvaluationImageUniforms = */ 0,
                    /* .MaxGeometryImageUniforms = */ 0,
                    /* .MaxFragmentImageUniforms = */ 8,
                    /* .MaxCombinedImageUniforms = */ 8,
                    /* .MaxGeometryTextureImageUnits = */ 16,
                    /* .MaxGeometryOutputVertices = */ 256,
                    /* .MaxGeometryTotalOutputComponents = */ 1024,
                    /* .MaxGeometryUniformComponents = */ 1024,
                    /* .MaxGeometryVaryingComponents = */ 64,
                    /* .MaxTessControlInputComponents = */ 128,
                    /* .MaxTessControlOutputComponents = */ 128,
                    /* .MaxTessControlTextureImageUnits = */ 16,
                    /* .MaxTessControlUniformComponents = */ 1024,
                    /* .MaxTessControlTotalOutputComponents = */ 4096,
                    /* .MaxTessEvaluationInputComponents = */ 128,
                    /* .MaxTessEvaluationOutputComponents = */ 128,
                    /* .MaxTessEvaluationTextureImageUnits = */ 16,
                    /* .MaxTessEvaluationUniformComponents = */ 1024,
                    /* .MaxTessPatchComponents = */ 120,
                    /* .MaxPatchVertices = */ 32,
                    /* .MaxTessGenLevel = */ 64,
                    /* .MaxViewports = */ 16,
                    /* .MaxVertexAtomicCounters = */ 0,
                    /* .MaxTessControlAtomicCounters = */ 0,
                    /* .MaxTessEvaluationAtomicCounters = */ 0,
                    /* .MaxGeometryAtomicCounters = */ 0,
                    /* .MaxFragmentAtomicCounters = */ 8,
                    /* .MaxCombinedAtomicCounters = */ 8,
                    /* .MaxAtomicCounterBindings = */ 1,
                    /* .MaxVertexAtomicCounterBuffers = */ 0,
                    /* .MaxTessControlAtomicCounterBuffers = */ 0,
                    /* .MaxTessEvaluationAtomicCounterBuffers = */ 0,
                    /* .MaxGeometryAtomicCounterBuffers = */ 0,
                    /* .MaxFragmentAtomicCounterBuffers = */ 1,
                    /* .MaxCombinedAtomicCounterBuffers = */ 1,
                    /* .MaxAtomicCounterBufferSize = */ 16384,
                    /* .MaxTransformFeedbackBuffers = */ 4,
                    /* .MaxTransformFeedbackInterleavedComponents = */ 64,
                    /* .MaxCullDistances = */ 8,
                    /* .MaxCombinedClipAndCullDistances = */ 8,
                    /* .MaxSamples = */ 4,
                    /* .maxMeshOutputVerticesNV = */ 256,
                    /* .maxMeshOutputPrimitivesNV = */ 512,
                    /* .maxMeshWorkGroupSizeX_NV = */ 32,
                    /* .maxMeshWorkGroupSizeY_NV = */ 1,
                    /* .maxMeshWorkGroupSizeZ_NV = */ 1,
                    /* .maxTaskWorkGroupSizeX_NV = */ 32,
                    /* .maxTaskWorkGroupSizeY_NV = */ 1,
                    /* .maxTaskWorkGroupSizeZ_NV = */ 1,
                    /* .maxMeshViewCountNV = */ 4,
                    /* .maxDualSourceDrawBuffersEXT = */ 1,

                    /* .limits = */{
                    /* .nonInductiveForLoops = */ 1,
                    /* .whileLoops = */ 1,
                    /* .doWhileLoops = */ 1,
                    /* .generalUniformIndexing = */ 1,
                    /* .generalAttributeMatrixVectorIndexing = */ 1,
                    /* .generalVaryingIndexing = */ 1,
                    /* .generalSamplerIndexing = */ 1,
                    /* .generalVariableIndexing = */ 1,
                    /* .generalConstantMatrixVectorIndexing = */ 1,
                }};
        }
        std::string source;
        std::string errors;
        std::string infos;
        std::vector<unsigned int> spirv;
        uint32_t result;
        glslang_input_t input;
        glslang_resource_t resources;
        glslang_shader_t* shader = nullptr;
        glslang_program_t* program = nullptr;
    };

    void GlslCompiler::Dispose() {

    }

    void GlslCompiler::Compile(vk::ShaderStageFlags stage, uint8_t* src, size_t src_len, void*& instance)
    {
        auto cp = new CompilePass();
        cp->source.append(reinterpret_cast<const char*>(src), src_len);
        cp->source.push_back(0);
        cp->input.code = cp->source.c_str();
        if (stage == vk::ShaderStageFlagBits::eFragment) {
            cp->input.stage = GLSLANG_STAGE_FRAGMENT;
        }
        else if (stage == vk::ShaderStageFlagBits::eVertex) {
            cp->input.stage = GLSLANG_STAGE_VERTEX;
        }
        else if (stage == vk::ShaderStageFlagBits::eGeometry) {
            cp->input.stage = GLSLANG_STAGE_GEOMETRY;
        }
        else {
            throw new std::runtime_error("Invalid stage");
        }
        fillLimits(cp);
        // cp->input.stage = stage;
        instance = cp;
        cp->result = 10;
        glslang_initialize_process();
        cp->shader = glslang_shader_create(&(cp->input));
        if (!glslang_shader_preprocess(cp->shader, &(cp->input))) {
            getErrors(cp);
            return;
        }
        cp->result = 11;
        if (!glslang_shader_parse(cp->shader, &(cp->input)))
        {
            getErrors(cp);
            return;
        }
        auto debugMsg = glslang_shader_get_info_debug_log(cp->shader);
        if (debugMsg != nullptr) {
            cp->infos.append(debugMsg);
        }
        cp->result = 12;
        cp->program = glslang_program_create();
        glslang_program_add_shader(cp->program, cp->shader);
        if (!glslang_program_link(cp->program, GLSLANG_MSG_SPV_RULES_BIT | GLSLANG_MSG_VULKAN_RULES_BIT))
        {
            getLinkErrors(cp);
            return;
        }
        debugMsg = glslang_program_get_info_debug_log(cp->program);
        if (debugMsg != nullptr) {
            cp->infos.append(debugMsg);
        }

        glslang_program_SPIRV_generate(cp->program, cp->input.stage);

        debugMsg = glslang_program_SPIRV_get_messages(cp->program);
        if (debugMsg != nullptr)
        {
            cp->result = 1;
            cp->infos.append(debugMsg);
        } else {
            cp->result = 0;
        }
        auto spvSize = glslang_program_SPIRV_get_size(cp->program);
        cp->spirv.resize(spvSize);
        std::copy_n(glslang_program_SPIRV_get_ptr(cp->program), spvSize, cp->spirv.data());
        glslang_finalize_process();
    }

    void GlslCompiler::GetOutput(void* instance, void*& msg, uint64_t& msg_len, uint32_t& result)
    {
        auto cp = reinterpret_cast<CompilePass*>(instance);
        result = cp->result;
        if (cp->result >= 10) {
            msg = &(cp->errors.at(0));
            msg_len = cp->errors.size();
        } else if (cp->infos.size() > 0) {
            msg = &(cp->infos.at(0));
            msg_len = cp->infos.size();
        }
        else {
            msg = nullptr;
            msg_len = 0;
        }
    }

    void GlslCompiler::GetSpirv(void* instance, void*& spirv, uint64_t& spirv_len)
    {
        auto cp = reinterpret_cast<CompilePass*>(instance);
        spirv = cp->spirv.data();
        spirv_len = sizeof(unsigned int) * cp->spirv.size();
    }

    void GlslCompiler::Free(void* instance)
    {
        if (instance == nullptr) {
            return;
        }
        auto cp = reinterpret_cast<CompilePass*>(instance);
        if (cp->shader != nullptr) {
            glslang_shader_delete(cp->shader);
        }
        if (cp->program != nullptr) {
            glslang_program_delete(cp->program);
        }
        delete cp;
    }

    void GlslCompiler::getErrors(CompilePass* cp)
    {
        auto msg = glslang_shader_get_info_log(cp->shader);
        cp->errors.append(msg);
    }

    void GlslCompiler::getLinkErrors(CompilePass* cp)
    {
        auto msg = glslang_program_get_info_log(cp->program);
        cp->errors.append(msg);
    }

    void GlslCompiler::fillLimits(CompilePass* cp)
    {
        auto &props = _dev->get_pdProperties();
        cp->resources.max_combined_image_units_and_fragment_outputs = props.limits.maxFragmentCombinedOutputResources;

    }
}