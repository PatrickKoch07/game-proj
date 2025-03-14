#version 410 core
// pos2D, tex2D
layout (location = 0) in vec4 vPos;

uniform mat4 transform;
uniform mat4 scale;
uniform mat4 projection;

out vec2 TexCoord;

void main()
{
    gl_Position = projection * (transform * (scale * vec4(vPos.x, vPos.y, vPos.y, 1.0f)));
    TexCoord = vPos.zw;
}