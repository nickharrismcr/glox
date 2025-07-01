#version 330

// Fragment shader with rainbow color effect based on incoming colors
in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec4 colDiffuse;
uniform float time;            // Time uniform for animation

out vec4 finalColor;

void main()
{
    // Use the incoming fragment color to influence the rainbow effect
    vec3 baseColor = fragColor.rgb * colDiffuse.rgb;
    
    // Calculate color shift based on the base color's hue and time
    // Convert RGB to a hue-like value for color shifting
    float colorIntensity = (baseColor.r + baseColor.g + baseColor.b) / 2.0;
    float colorHue = atan(baseColor.g - baseColor.b, baseColor.r - baseColor.g) + time * 2.0;
    
    // Create rainbow shifts that are influenced by the original color
    vec3 rainbow = vec3(
        0.5 + 0.5 * sin(colorHue + time * 1.5),
        0.5 + 0.5 * sin(colorHue + time * 1.5 + 2.094),
        0.5 + 0.5 * sin(colorHue + time * 1.5 + 4.188)
    );
    
     
    // Sample the texture
    vec4 texelColor = texture(texture0, fragTexCoord);
    
    // Blend the original color with the rainbow effect
    // The rainbow effect is modulated by the original color intensity
    vec3 finalRGB = rainbow * colorIntensity * texelColor.rgb;
    
    finalColor = vec4(finalRGB, fragColor.a * colDiffuse.a);
}
