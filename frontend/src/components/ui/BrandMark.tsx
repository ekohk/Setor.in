/**
 * Brand mark — leaf-in-rounded-square, identical to the one in email/base.html
 * so the visual identity stays consistent across web + email.
 */
export function BrandMark({ size = 28 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 28 28"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Setor.in"
    >
      <rect width="28" height="28" rx="8" fill="#1f5d3a" />
      <path
        d="M21 9c0 6.5-4.5 10.5-9 10.5-1.5 0-3-.6-3-.6s2.7-.5 4.5-3.3c-1.5.7-3.3.7-3.3.7s1.4-2.2 4-3.3c-2.2 0-3.3.7-3.3.7s2.7-3.6 10.1-4.7z"
        fill="#ffffff"
      />
    </svg>
  );
}
