"""Float16 embedding storage helpers for Zep integrations.

Wire format mirrors pkg/embeddings (ZEP1):
  magic[4]='ZEP1', version=1, dtype=1 (float16) or 2 (float32 legacy),
  ndims uint32 LE, payload ndims * 2 (or * 4 for legacy) bytes LE.
"""

from __future__ import annotations

import struct
from array import array
from typing import Iterable, List, Sequence, Union

MAGIC = b"ZEP1"
FORMAT_VERSION = 1
DTYPE_FLOAT16 = 1
DTYPE_FLOAT32 = 2
_HEADER = struct.Struct("<4sBBI")  # magic, version, dtype, ndims


class Float16Embedding:
    """Embedding held as IEEE-754 binary16 bit patterns (array of uint16)."""

    __slots__ = ("_bits",)

    def __init__(self, bits: Sequence[int] | array):
        if isinstance(bits, array):
            self._bits = bits
        else:
            self._bits = array("H", bits)

    @classmethod
    def from_float32(cls, values: Sequence[float]) -> "Float16Embedding":
        # Use struct for f32->f16 via numpy-free path: pack as f32 and use
        # Python's struct doesn't do f16 on all platforms; implement via
        # bit manipulation compatible with Go codec.
        bits = array("H", (float32_to_float16bits(float(v)) for v in values))
        return cls(bits)

    def to_float32(self) -> List[float]:
        return [float16bits_to_float32(b) for b in self._bits]

    @property
    def dims(self) -> int:
        return len(self._bits)

    def byte_size(self) -> int:
        return len(self._bits) * 2

    def encode(self) -> bytes:
        if not self._bits:
            raise ValueError("empty vector")
        header = _HEADER.pack(MAGIC, FORMAT_VERSION, DTYPE_FLOAT16, len(self._bits))
        return header + self._bits.tobytes()

    @classmethod
    def decode(cls, blob: bytes) -> "Float16Embedding":
        if len(blob) < 10:
            # try raw float32 LE
            if len(blob) >= 4 and len(blob) % 4 == 0:
                n = len(blob) // 4
                vals = struct.unpack(f"<{n}f", blob)
                return cls.from_float32(vals)
            raise ValueError("truncated embedding blob")
        magic, version, dtype, ndims = _HEADER.unpack_from(blob, 0)
        if magic != MAGIC:
            if len(blob) % 4 == 0:
                n = len(blob) // 4
                vals = struct.unpack(f"<{n}f", blob)
                return cls.from_float32(vals)
            raise ValueError("bad magic")
        if version != FORMAT_VERSION:
            raise ValueError(f"unsupported version {version}")
        if dtype == DTYPE_FLOAT16:
            need = 10 + ndims * 2
            if len(blob) < need:
                raise ValueError("truncated float16 payload")
            bits = array("H")
            bits.frombytes(blob[10:need])
            return cls(bits)
        if dtype == DTYPE_FLOAT32:
            need = 10 + ndims * 4
            if len(blob) < need:
                raise ValueError("truncated float32 payload")
            vals = struct.unpack(f"<{ndims}f", blob[10:need])
            return cls.from_float32(vals)
        raise ValueError(f"unknown dtype {dtype}")


def float32_to_float16bits(f: float) -> int:
    """IEEE-754 binary16 bits from Python float (via float32)."""
    u = struct.unpack(">I", struct.pack(">f", f))[0]
    sign = (u >> 16) & 0x8000
    exp = ((u >> 23) & 0xFF) - 127 + 15
    mant = u & 0x7FFFFF
    if ((u >> 23) & 0xFF) == 0xFF:
        return sign | (0x7E00 if mant else 0x7C00)
    if exp >= 0x1F:
        return sign | 0x7C00
    if exp <= 0:
        if exp < -10:
            return sign
        mant |= 0x800000
        shift = 14 - exp
        mant = (mant + (1 << (shift - 1))) >> shift
        return sign | mant
    rounded = mant + 0x1000
    if rounded & 0x800000:
        exp += 1
        rounded = 0
        if exp >= 0x1F:
            return sign | 0x7C00
    return sign | (exp << 10) | (rounded >> 13)


def float16bits_to_float32(h: int) -> float:
    sign = (h & 0x8000) << 16
    exp = (h >> 10) & 0x1F
    mant = h & 0x3FF
    if exp == 0:
        if mant == 0:
            return struct.unpack(">f", struct.pack(">I", sign))[0]
        exp32 = 127 - 15 + 1
        while (mant & 0x400) == 0:
            mant <<= 1
            exp32 -= 1
        mant &= 0x3FF
        bits = sign | (exp32 << 23) | (mant << 13)
        return struct.unpack(">f", struct.pack(">I", bits))[0]
    if exp == 0x1F:
        bits = sign | (0x7F800000 if mant == 0 else 0x7FC00000)
        return struct.unpack(">f", struct.pack(">I", bits))[0]
    bits = sign | ((exp - 15 + 127) << 23) | (mant << 13)
    return struct.unpack(">f", struct.pack(">I", bits))[0]


def is_float16_blob(blob: bytes) -> bool:
    return (
        len(blob) >= 10
        and blob[:4] == MAGIC
        and blob[4] == FORMAT_VERSION
        and blob[5] == DTYPE_FLOAT16
    )


__all__ = [
    "Float16Embedding",
    "float32_to_float16bits",
    "float16bits_to_float32",
    "is_float16_blob",
    "MAGIC",
    "DTYPE_FLOAT16",
    "DTYPE_FLOAT32",
]
