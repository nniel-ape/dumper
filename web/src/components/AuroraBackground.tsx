export function AuroraBackground() {
  return (
    <div className="fixed inset-0 -z-10 overflow-hidden pointer-events-none">
      {/* Gradient orb 1 - Indigo */}
      <div
        className="absolute -top-40 -left-40 h-80 w-80 rounded-full aurora-orb-1 blur-3xl animate-aurora-float will-change-blur"
        aria-hidden="true"
      />
      {/* Gradient orb 2 - Violet */}
      <div
        className="absolute top-1/3 -right-20 h-96 w-96 rounded-full aurora-orb-2 blur-3xl animate-aurora-float-delayed will-change-blur"
        aria-hidden="true"
      />
      {/* Gradient orb 3 - Pink */}
      <div
        className="absolute -bottom-20 left-1/3 h-72 w-72 rounded-full aurora-orb-3 blur-3xl animate-aurora-float-slow will-change-blur"
        aria-hidden="true"
      />
    </div>
  )
}
