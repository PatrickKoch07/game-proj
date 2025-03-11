#version 410 core
// pos2D, tex2D
layout (location = 0) in vec4 vPos;

uniform mat3 transform;
unfirom mat4 projection;

outt vec2 TexCoord;

void main()
{
    gl_Position = transform * vec4(vPos.x, vPos.y, 0.0, -1.0 * (vPos.y-2.0));
    TexCoord = vPos.zw;
}