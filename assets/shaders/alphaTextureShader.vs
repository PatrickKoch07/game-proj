#version 410 core
// pos2D, tex2D
layout (location = 0) in vec4 vPos;

uniform mat4 transform;
uniform mat4 scale;
uniform mat4 projection;

out vec2 TexCoord;

void main()
{
    gl_Position = projection * (transform * (scale * vec4(vPos.x, vPos.y, 1.0, 1.0f)));
    gl_Position.z = gl_Position.z * 0.8f + 0.2f;  // moves range of -1,1 to -.6,1
    TexCoord = vPos.zw;
}