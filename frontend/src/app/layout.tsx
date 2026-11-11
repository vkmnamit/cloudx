import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = { title: "CloudX — Ship code. Skip the complexity.", description: "Deploy and manage your applications with CloudX." };

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="en"><head><link rel="preconnect" href="https://fonts.googleapis.com"/><link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous"/><link href="https://fonts.googleapis.com/css2?family=Anton&display=swap" rel="stylesheet"/></head><body>{children}</body></html>;
}
