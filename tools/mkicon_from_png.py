"""Tạo assets/app.ico từ ảnh PNG vuông: làm trong suốt 4 góc trắng, xóa dấu ✦ ở góc dưới phải."""
from PIL import Image, ImageDraw
import sys
src, out = sys.argv[1], sys.argv[2]
im = Image.open(src).convert('RGBA')
W, H = im.size
d = ImageDraw.Draw(im)
# Xóa dấu ✦ (vùng nhỏ góc dưới phải) bằng màu nền
bg = im.getpixel((int(W*0.84), int(H*0.86)))
d.ellipse((int(W*0.84), int(H*0.845), int(W*0.925), int(H*0.925)), fill=bg)
# Góc trắng → trong suốt (loang từ 4 góc)
marker = (1, 2, 3, 0)
for c in [(0, 0), (W-1, 0), (0, H-1), (W-1, H-1)]:
    ImageDraw.floodfill(im, c, marker, thresh=60)
px = im.load()
for y in range(H):
    for x in range(W):
        if px[x, y] == marker:
            px[x, y] = (0, 0, 0, 0)
# Cắt sát nội dung và đưa về vuông
bbox = im.getbbox()
im = im.crop(bbox)
s = max(im.size)
sq = Image.new('RGBA', (s, s), (0, 0, 0, 0))
sq.paste(im, ((s - im.size[0]) // 2, (s - im.size[1]) // 2))
big = sq.resize((256, 256), Image.LANCZOS)
big.save(out, sizes=[(16, 16), (20, 20), (24, 24), (32, 32), (40, 40), (48, 48), (64, 64), (128, 128), (256, 256)])
big.save(out.replace('.ico', '.png'))
sq.resize((512, 512), Image.LANCZOS).save(out.replace('.ico', '_512.png'))
