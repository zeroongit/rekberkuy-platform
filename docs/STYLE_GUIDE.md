# Documentation Style Guide

## Bahasa

- **User guide** → Bahasa Indonesia, ramah, hindari jargon.
- **Dokumen teknis** → Bahasa Inggris, ringkas, presisi.
- **Istilah domain** (Rekber, RekberPay, KTP/SIM) → tetap Bahasa Indonesia, italicize saat pertama muncul: _Rekber_.

## Struktur File

Setiap file wajib punya:

```markdown
# <Judul>

> **Terakhir diverifikasi:** YYYY-MM-DD
> **Audiens:** <end user | engineer | integrator>

<konten>

# Heading

1. hanya satu per file (judul).
2. untuk section utama — wajib masuk TOC.
3.  untuk sub-section.

Jangan lompat level (# → ### ❌).

Link
Link internal: [teks](./relative/path.md) — pakai path relatif.

Anchor internal: [teks](#anchor-name) — JANGAN pakai https://github.com#....

Link eksternal: [teks](https://...).

Tabel
Selalu pakai header row.

Alignment eksplisit untuk kolom angka: |---:|.

Commit Message
Format: <scope>: <deskripsi imperatif>

Contoh:

docs: add README as entry point

docs(adr): add template

docs(api): document environment variables