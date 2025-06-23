'use client';

import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/hooks/useAuth";
import { AlertProvider } from "@/contexts/AlertContext";
import { ToastProvider } from "@/contexts/ToastContext";
import Navbar from "@/components/Navbar";
import ToastContainer from "@/components/Toast";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <head>
        <title>Social Network</title>
        <meta name="description" content="A Facebook-like social network application" />
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        <AuthProvider>
          <AlertProvider>
            <ToastProvider>
              <Navbar />
              <main>
                {children}
              </main>
              <ToastContainer />
            </ToastProvider>
          </AlertProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
