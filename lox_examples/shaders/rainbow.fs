#version 330

// Fragment shader with rainbow color effect
in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec4 colDiffuse;
uniform float time;            // Time uniform for animation

out vec4 finalColor;

void main()
{
    // Create rainbow colors based on time and position
    vec3 rainbow = vec3(
        0.5 + 0.5 * sin(time * 2.0 + fragTexCoord.x * 10.0),
        0.5 + 0.5 * sin(time * 2.0 + fragTexCoord.x * 10.0 + 2.094),
        0.5 + 0.5 * sin(time * 2.0 + fragTexCoord.x * 10.0 + 4.188)
    );
    
    // Sample the texture
    vec4 texelColor = texture(texture0, fragTexCoord);
    
    // Mix original color with rainbow effect
    vec3 finalRGB = mix(fragColor.rgb * colDiffuse.rgb, rainbow, 0.3);
    
    finalColor = vec4(finalRGB, fragColor.a * colDiffuse.a);
}
