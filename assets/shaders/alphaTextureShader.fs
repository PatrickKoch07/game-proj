#version 410 core
in vec2 TexCoord;

uniform sampler2D tex;

out vec4 FragColor;

void main()
{
    vec4 texColor = texture(tex, TexCoord);
    if (texColor.a <= 0.01)
        discard;
    FragColor = texColor;
}