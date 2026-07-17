with open('server.go','r') as f:
    lines = f.readlines()
for i in range(424, 470):
    if lines[i].strip():
        lines[i] = '\t' + lines[i]
with open('server.go','w') as f:
    f.writelines(lines)
