'use client';
import {Inter} from 'next/font/google';
import './globals.css';
import {Provider} from 'react-redux';
import store from '../store';
import Head from "next/head";

const inter = Inter({subsets: ['latin']});

const metadata = {
  title: 'AIEcho', description: 'Frontend for the AIEcho Accessibility Website'
};

export default function RootLayout({children}) {
  return (<Provider store={store}>
    <html lang="en">
    <head>
      <title>AIEcho Visual Assistant</title>
      <meta name="description" content="AI-based environment description service - 'be my eyes'. Free & open source!"/>
      <meta name="viewport" content="width=device-width, initial-scale=1"/>
      <meta name="application-name" content="AIEcho"/>
      <meta name="apple-mobile-web-app-title" content="AIEcho"/>
      <meta name="apple-mobile-web-app-capable" content="yes"/>
      <meta name="mobile-web-app-capable" content="yes"/>
      <meta name="apple-mobile-web-app-status-bar-style" content="black"/>
      <link rel="apple-touch-icon" href="/logo.png"/>
      <meta name="apple-mobile-web-app-status-bar-style" content="black"/>
    </head>
    <body className={inter.className}>{children}</body>
    </html>
  </Provider>);
}
