import "./globals.css";

export const metadata = {
  title: "Testra",
  description: "One Platform. Every Test.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
