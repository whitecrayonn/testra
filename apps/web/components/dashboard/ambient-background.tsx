export function AmbientBackground() {
  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 -z-10 transition-[background] duration-500 ease-out"
      style={{
        background:
          "radial-gradient(1100px 720px at 10% -10%, var(--blob1), transparent 62%), " +
          "radial-gradient(900px 640px at 94% 4%, var(--blob2), transparent 64%), " +
          "radial-gradient(820px 620px at 58% 112%, var(--blob3), transparent 62%), " +
          "var(--bg)",
      }}
    />
  );
}
