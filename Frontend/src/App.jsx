import React from "react";
import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext.jsx';
import { SplashPage } from './pages/SplashPage.jsx';
import { RolePage } from './pages/RolePage.jsx';
import { ProfileSetupPage } from './pages/ProfileSetupPage.jsx';
import { NibPage } from './pages/NibPage.jsx';
import { HomePage } from './pages/HomePage.jsx';
import { PostFormPage } from './pages/PostFormPage.jsx';
import { MatchPage } from './pages/MatchPage.jsx';
import { NotificationsPage } from './pages/NotificationsPage.jsx';
import { ChatPage } from './pages/ChatPage.jsx';
import { AgreementPage } from './pages/AgreementPage.jsx';
import { RfqPage } from './pages/RfqPage.jsx';
import { ContactPage } from './pages/ContactPage.jsx';
import { ProfilePage } from './pages/ProfilePage.jsx';
import "./styles.css";

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<SplashPage />} />
        <Route path="/role" element={<RolePage />} />
        <Route path="/profile-setup" element={<ProfileSetupPage />} />
        <Route path="/nib" element={<NibPage />} />
        <Route path="/home" element={<HomePage />} />
        <Route path="/post" element={<PostFormPage />} />
        <Route path="/match" element={<MatchPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route path="/chat" element={<ChatPage />} />
        <Route path="/agreement" element={<AgreementPage />} />
        <Route path="/rfq" element={<RfqPage />} />
        <Route path="/contact" element={<ContactPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="*" element={<Navigate to="/home" replace />} />
      </Routes>
    </AuthProvider>
  );
}
