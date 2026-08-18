import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 p-6">
      <div className="max-w-md w-full bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-xl p-8 text-center">
        <div className="w-16 h-16 bg-blue-100 dark:bg-blue-950 text-blue-600 dark:text-blue-400 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl font-bold">
          404
        </div>
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100 mb-2">
          Halaman Tidak Ditemukan
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-6">
          Halaman yang Anda tuju mungkin telah dihapus, berganti nama, atau untuk sementara tidak tersedia.
        </p>
        <Link
          href="/"
          className="inline-block px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl text-sm transition shadow-md"
        >
          Kembali ke Beranda
        </Link>
      </div>
    </div>
  );
}
