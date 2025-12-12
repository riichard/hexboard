import sys
W,H=1280,720
for y in range(H):
  for x in range(W):
    sys.stdout.buffer.write(bytes([x*255//(W-1)]))
