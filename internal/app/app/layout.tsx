import type { Metadata } from "next";
import { Inter, Geist } from "next/font/google";
import "../styles/globals.css";
import { cn } from "@/lib/utils";

const geist = Geist({ subsets: ["latin"], variable: "--font-sans" });

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "ZyrisRAG | Multimodal AI Brain",
  description:
    "Deep technical memory for integration with Godot and Zyris Engine.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR" className={cn("dark", "font-sans", geist.variable)}>
      <body
        className={`${inter.className} bg-background text-foreground antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
