import { DashboardPage } from "../features/dashboard/DashboardPage.jsx";
import { AuthPage } from "../features/auth/AuthPage.jsx";
import { useAuth } from "../features/auth/hooks/useAuth.js";

export function App() {
  const auth = useAuth();

  if (!auth.token) {
    return <AuthPage onLogin={auth.login} onRegister={auth.register} isLoading={auth.isLoading} error={auth.error} />;
  }

  return (
    <DashboardPage
      authToken={auth.token}
      currentUser={auth.user}
      onLogout={auth.logout}
      onUpdateProfile={auth.updateProfile}
    />
  );
}
