# Rekberkuy Platform — Documentation

Kumpulan dokumentasi untuk platform Rekberkuy. Mulai dari sini.

## 📚 Daftar Dokumen

| Dokumen | Audiens | Deskripsi |
|---|---|---|
| [General User Guide](./general-user-guide.md) | End user (buyer/seller) | Panduan penggunaan fitur: wallet, goods, services, events |
| [Application Flow & Module Guide](./application-flow-and-module-guide.md) | Engineer, PM | Alur transaksi, state machine, money flow, peta modul |
| [ADR](./adr/) | Engineer, arsitek | Architecture Decision Records — kenapa keputusan teknis diambil |
| [Agents](./agents/) | AI agent, kontributor | Panduan domain, issue tracker, triage labels |
| [API Collection](./api/) | Developer, integrator | Postman collection untuk core service |

## 🧭 Urutan Baca yang Disarankan

**Kalau kamu end user:**
1. [General User Guide](./general-user-guide.md)

**Kalau kamu engineer baru:**
1. [Application Flow & Module Guide](./application-flow-and-module-guide.md)
2. [ADR](./adr/) — baca dari nomor terkecil
3. [Agents/domain.md](./agents/domain.md)

**Kalau kamu integrator API:**
1. [API Collection](./api/)
2. [Application Flow & Module Guide](./application-flow-and-module-guide.md) — bagian money flow

## 🛠️ Berkontribusi

- Gunakan **Bahasa Inggris** untuk dokumen teknis, **Bahasa Indonesia** untuk user guide.
- Setiap file baru wajib punya header metadata (lihat [style guide](./STYLE_GUIDE.md)).
- Keputusan teknis besar → buat ADR baru di [`adr/`](./adr/) dengan nomor urut berikutnya.
- Update TOC setiap kali menambah heading `##` atau `###`.

## 📋 Status Dokumen

| Dokumen | Terakhir Diverifikasi |
|---|---|
| general-user-guide.md | _isi tanggal_ |
| application-flow-and-module-guide.md | _isi tanggal_ |
| adr/ | _isi tanggal_ |
| agents/ | _isi tanggal_ |
| api/ | _isi tanggal_ |