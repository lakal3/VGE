
const float shadowBias = 0.005;

void shadowFactorFormula() {
#if probe
    l.shadowFactor = 1.0;
#endif
#if !probe
    if (l.tx_shadowmap != 0) {
        if (l.lenToPos > l.maxDistance) {
            l.shadowFactor = 0;
        } else {
            vec3 samplePos = l.lightToPos / vec3(-l.lenToPos);
            float depth = 0;
            vec2 offset = vec2(0);
            switch (l.kind) {
            case 2:// Spot light
                float n = samplePos.y + 1;
                vec2 pos = vec2(samplePos.x / n, samplePos.z / n) * vec2(0.5, 0.5) + vec2(0.5, 0.5) + offset;
                depth = texture(frameImages2D[l.tx_shadowmap], pos).x;
                break;
            case 1:
                if (samplePos.y < 0) {
                    float n = -samplePos.y + 1;
                    vec3 pos = vec3(samplePos.x / n, samplePos.z / n, 0) * vec3(0.5, 0.5, 1) + vec3(0.5, 0.5, 0) + vec3(offset, 0);
                    depth = texture(frameImagesArray[l.tx_shadowmap], pos).x;
                } else {
                    float n = samplePos.y + 1;
                    vec3 pos = vec3(samplePos.x / n, samplePos.z / n, 1) * vec3(0.5, 0.5, 1) + vec3(0.5, 0.5, 0) + vec3(offset, 0);
                    depth = texture(frameImagesArray[l.tx_shadowmap], pos).x;
                }
            }
            l.shadowFactor -= (depth + shadowBias > (l.lenToPos / l.maxDistance)) ? 0.0 : 1.0;
        }
    }
#endif
}