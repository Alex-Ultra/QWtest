import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppRoot } from '@telegram-apps/telegram-ui';

import { Onboarding } from './pages/Onboarding';
import { Waiting } from './pages/Waiting';
import { Rejected } from './pages/Rejected';
import { Dashboard } from './pages/Dashboard';
import { TicketDetail } from './pages/TicketDetail';
import { CreateTicket } from './pages/CreateTicket';
import { Profile } from './pages/Profile';

function App() {
  return (
    <AppRoot>
      <Router>
        <Routes>
          <Route path="/" element={<Onboarding />} />
          <Route path="/waiting" element={<Waiting />} />
          <Route path="/rejected" element={<Rejected />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/ticket/:id" element={<TicketDetail />} />
          <Route path="/create-ticket" element={<CreateTicket />} />
          <Route path="/profile" element={<Profile />} />
        </Routes>
      </Router>
    </AppRoot>
  );
}

export default App;