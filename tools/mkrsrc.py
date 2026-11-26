"""Sinh file tài nguyên Windows (.syso, COFF amd64) gồm: icon, manifest, thông tin phiên bản.
Không cần windres/rsrc. Dùng: python3 mkrsrc.py app.ico VERSION "Mô tả" "TenFile.exe" out.syso
"""
import struct, sys

def u16z(s): return s.encode('utf-16-le') + b'\0\0'
def pad4(b): return b + b'\0' * ((4 - len(b) % 4) % 4)

def vnode(key, value=b'', wtype=0, children=(), vlen=None):
    body = struct.pack('<HHH', 0, 0, wtype) + u16z(key)
    body = pad4(body) + value
    for c in children:
        body = pad4(body) + c
    if vlen is None:
        vlen = len(value) // 2 if wtype == 1 else len(value)
    return struct.pack('<HHH', len(body), vlen, wtype) + body[6:]

def vstring(k, v):
    return vnode(k, u16z(v), 1)

def version_info(ver, desc, fname, product, company, copyright_):
    p = [int(x) for x in ver.split('.')] + [0] * 4
    ms, ls = (p[0] << 16) | p[1], (p[2] << 16) | p[3]
    fixed = struct.pack('<13I', 0xFEEF04BD, 0x00010000, ms, ls, ms, ls, 0x3F, 0, 0x40004, 1, 0, 0, 0)
    vs = '%d.%d.%d.%d' % tuple(p[:4])
    strs = [vstring(k, v) for k, v in [
        ('CompanyName', company), ('FileDescription', desc), ('FileVersion', vs),
        ('InternalName', fname.rsplit('.', 1)[0]), ('LegalCopyright', copyright_),
        ('OriginalFilename', fname), ('ProductName', product), ('ProductVersion', ver)]]
    st = vnode('040904B0', b'', 1, strs, vlen=0)
    sfi = vnode('StringFileInfo', b'', 1, [st], vlen=0)
    var = vnode('Translation', struct.pack('<HH', 0x0409, 1200), 0)
    vfi = vnode('VarFileInfo', b'', 1, [var], vlen=0)
    return vnode('VS_VERSION_INFO', fixed, 0, [sfi, vfi])

MANIFEST = b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
  <assemblyIdentity type="win32" name="DinhDangVanBanChuan" version="%s" processorArchitecture="*"/>
  <dependency><dependentAssembly>
    <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
  </dependentAssembly></dependency>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3"><security><requestedPrivileges>
    <requestedExecutionLevel level="asInvoker" uiAccess="false"/>
  </requestedPrivileges></security></trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1"><application>
    <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
    <supportedOS Id="{1f676c76-80e1-4239-95bb-83d0f6d0da78}"/>
    <supportedOS Id="{4a2f28e3-53b9-4441-ba9c-d69d4a4a6e38}"/>
    <supportedOS Id="{35138b9a-5d96-4fbd-8e2d-a2440225f93a}"/>
  </application></compatibility>
  <asmv3:application><asmv3:windowsSettings xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">
    <dpiAware>true</dpiAware>
  </asmv3:windowsSettings></asmv3:application>
</assembly>
'''

def read_ico(path):
    b = open(path, 'rb').read()
    _, _, n = struct.unpack('<HHH', b[:6])
    entries = []
    for i in range(n):
        w, h, cc, r, planes, bpp, size, off = struct.unpack('<BBBBHHII', b[6 + 16 * i: 22 + 16 * i])
        entries.append(((w, h, cc, r, planes, bpp), b[off:off + size]))
    return entries

def build(ico, ver, desc, fname, out):
    icons = read_ico(ico)
    res = {}  # type -> {id: data}
    res[3] = {i + 1: data for i, (_, data) in enumerate(icons)}
    grp = struct.pack('<HHH', 0, 1, len(icons))
    for i, ((w, h, cc, r, planes, bpp), data) in enumerate(icons):
        grp += struct.pack('<BBBBHHIH', w, h, cc, r, planes or 1, bpp or 32, len(data), i + 1)
    res[14] = {1: grp}
    res[16] = {1: version_info(ver, desc, fname, 'ĐỊNH DẠNG VĂN BẢN CHUẨN',
                               'Phòng Kỹ thuật - VTTB - CNTT, Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2',
                               '© 2026 Hồ Viết Thắng')}
    res[24] = {1: MANIFEST.replace(b'%s', ('%s.0' % ver).encode())}

    types = sorted(res)
    # Tính kích thước các bảng thư mục
    def dirsize(n): return 16 + 8 * n
    off = dirsize(len(types))
    l2_off = {}
    for t in types:
        l2_off[t] = off; off += dirsize(len(res[t]))
    l3_off = {}
    for t in types:
        for i in sorted(res[t]):
            l3_off[(t, i)] = off; off += dirsize(1)
    de_off = {}
    for t in types:
        for i in sorted(res[t]):
            de_off[(t, i)] = off; off += 16
    data_off = {}
    for t in types:
        for i in sorted(res[t]):
            off = (off + 7) & ~7
            data_off[(t, i)] = off; off += len(res[t][i])
    raw = bytearray(off)
    def put_dir(at, entries):
        struct.pack_into('<IIHHHH', raw, at, 0, 0, 0, 0, 0, len(entries))
        for k, (name, ptr) in enumerate(entries):
            struct.pack_into('<II', raw, at + 16 + 8 * k, name, ptr)
    put_dir(0, [(t, 0x80000000 | l2_off[t]) for t in types])
    relocs = []
    for t in types:
        put_dir(l2_off[t], [(i, 0x80000000 | l3_off[(t, i)]) for i in sorted(res[t])])
        for i in sorted(res[t]):
            put_dir(l3_off[(t, i)], [(0x0409, de_off[(t, i)])])
            d = res[t][i]
            struct.pack_into('<IIII', raw, de_off[(t, i)], data_off[(t, i)], len(d), 0, 0)
            relocs.append(de_off[(t, i)])
            raw[data_off[(t, i)]:data_off[(t, i)] + len(d)] = d
    raw = bytes(raw)
    raw += b'\0' * ((4 - len(raw) % 4) % 4)
    hdr_size, sec_size = 20, 40
    p_raw = hdr_size + sec_size
    p_rel = p_raw + len(raw)
    p_sym = p_rel + 10 * len(relocs)
    coff = struct.pack('<HHIIIHH', 0x8664, 1, 0, p_sym, 1, 0, 0)
    sec = struct.pack('<8sIIIIIIHHI', b'.rsrc', 0, 0, len(raw), p_raw, p_rel, 0, len(relocs), 0, 0x40000040)
    rel = b''.join(struct.pack('<IIH', r, 0, 3) for r in relocs)  # IMAGE_REL_AMD64_ADDR32NB
    sym = struct.pack('<8sIhHBB', b'.rsrc', 0, 1, 0, 3, 0)
    open(out, 'wb').write(coff + sec + raw + rel + sym + struct.pack('<I', 4))

if __name__ == '__main__':
    build(*sys.argv[1:6])
