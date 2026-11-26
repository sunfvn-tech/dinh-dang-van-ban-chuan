from PIL import Image, ImageDraw, ImageFont
import sys
S=256
im=Image.new('RGBA',(S,S),(0,0,0,0)); d=ImageDraw.Draw(im)
# trang giấy
d.rounded_rectangle((40,14,200,236),radius=14,fill=(255,255,255,255),outline=(40,70,140,255),width=8)
d.polygon([(160,14),(200,54),(160,54)],fill=(200,215,240,255))
d.line([(160,14),(160,54),(200,54)],fill=(40,70,140,255),width=6)
# dải tiêu đề
d.rectangle((62,66,178,82),fill=(30,80,170,255))
d.rectangle((62,94,150,104),fill=(30,80,170,255))
for y in (122,142,162,182):
    d.rectangle((62,y,178,y+8),fill=(150,160,180,255))
# dấu kiểm xanh
d.ellipse((132,150,244,252),fill=(20,150,70,255),outline=(255,255,255,255),width=6)
d.line([(158,202),(182,226),(220,176)],fill=(255,255,255,255),width=16,joint='curve')
im.save(sys.argv[1],sizes=[(16,16),(24,24),(32,32),(48,48),(64,64),(128,128),(256,256)])
im.save(sys.argv[1].replace('.ico','.png'))
