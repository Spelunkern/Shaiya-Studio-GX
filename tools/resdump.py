#!/usr/bin/env python3
"""Dump menus, accelerators, dialogs and string tables of shStudio's
compiled res/shstudio.res as text (used to port the UI)."""
import struct, sys

def u16s(b, o):
    s = []
    while True:
        c = struct.unpack_from('<H', b, o)[0]; o += 2
        if c == 0: return ''.join(s), o
        s.append(chr(c))

def name_or_id(b, o):
    v = struct.unpack_from('<H', b, o)[0]
    if v == 0xFFFF:
        return struct.unpack_from('<H', b, o + 2)[0], o + 4
    return u16s(b, o)

def entries(data):
    o = 0
    while o + 32 <= len(data):
        size, hsize = struct.unpack_from('<II', data, o)
        t, p = name_or_id(data, o + 8)
        n, p = name_or_id(data, p)
        body = data[o + hsize:o + hsize + size]
        yield t, n, body
        o += hsize + size
        o = (o + 3) & ~3

def menu(b):
    ver, off = struct.unpack_from('<HH', b, 0)
    o = 4 + off
    out = []; depth = 0
    stack = []
    if ver == 1:
        out.append('(extended menu)'); return out
    while o < len(b):
        flags = struct.unpack_from('<H', b, o)[0]; o += 2
        if flags & 0x10:
            text, o = u16s(b, o)
            out.append('  ' * len(stack) + f'POPUP "{text}"')
            stack.append(flags & 0x80)
            continue
        mid = struct.unpack_from('<H', b, o)[0]; o += 2
        text, o = u16s(b, o)
        extra = []
        if flags & 1: extra.append('GRAYED')
        if flags & 8: extra.append('CHECKED')
        if flags & 0x800 or (mid == 0 and text == ''): out.append('  ' * len(stack) + 'SEPARATOR')
        else: out.append('  ' * len(stack) + f'MENUITEM "{text}" 0x{mid:04X} {" ".join(extra)}')
        if flags & 0x80:
            while stack:
                last = stack.pop()
                if not last: break
            if not stack and o < len(b) and False: break
    return out

def accel(b):
    out = []
    for i in range(0, len(b) - 7, 8):
        f, key, cmd = struct.unpack_from('<HHH', b, i)
        mods = [m for bit, m in ((4, 'Shift'), (8, 'Ctrl'), (16, 'Alt')) if f & bit]
        k = f'VK_{key:02X}' if f & 1 else repr(chr(key))
        out.append(f'  {"+".join(mods + [k])} -> 0x{cmd:04X}')
    return out

def strings(n, b):
    out = []; o = 0
    for i in range(16):
        ln = struct.unpack_from('<H', b, o)[0]; o += 2
        s = b[o:o + 2 * ln].decode('utf-16-le'); o += 2 * ln
        if ln: out.append(f'  {(n - 1) * 16 + i}: {s!r}')
    return out

def dialog(b):
    sig = struct.unpack_from('<HH', b, 0)
    out = []
    if sig[1] == 0xFFFF:
        out.append('  (DIALOGEX)')
        o = 26
        o = name_or_id(b, o)[1]; o = name_or_id(b, o)[1]
        title, o = u16s(b, o)
        out.append(f'  title {title!r}')
        return out
    return ['  (DIALOG)']

data = open(sys.argv[1], 'rb').read()
for t, n, body in entries(data):
    if t == 4:
        print(f'MENU {n}'); print('\n'.join(menu(body)))
    elif t == 9:
        print(f'ACCELERATORS {n}'); print('\n'.join(accel(body)))
    elif t == 6:
        print(f'STRINGTABLE {n}'); print('\n'.join(strings(n, body)))
    elif t == 5:
        print(f'DIALOG {n}'); print('\n'.join(dialog(body)))
