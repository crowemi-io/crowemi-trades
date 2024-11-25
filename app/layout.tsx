import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "crowemi-trades",
  description: "dater on actions taken by crowemi-trades.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        {children}
      </body>
    </html>
  );
}
