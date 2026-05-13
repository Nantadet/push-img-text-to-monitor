export default function HomePage() {
  return (
    <main className="max-w-3xl mx-auto p-8 space-y-8">
      <header>
        <h1 className="text-3xl font-semibold">HLLC Day 3</h1>
        <p className="text-slate-600">Next.js + Go + MongoDB · CRUD Majors & Courses</p>
      </header>

      <div className="bg-white rounded-xl p-5 shadow-sm space-y-3">
        <h2 className="font-semibold">หน้าแรก</h2>
        <p>ไปที่ <a href="/majors" className="text-blue-600 underline">/majors</a> เพื่อจัดการ Major</p>
      </div>
    </main>
  )
}
