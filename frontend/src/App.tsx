import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/AppShell";
import { ServiceCreatePage } from "./pages/ServiceCreatePage";
import { ServiceDetailPage } from "./pages/ServiceDetailPage";
import { ServiceEditPage } from "./pages/ServiceEditPage";
import { ServiceListPage } from "./pages/ServiceListPage";

export function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/services" replace />} />
        <Route path="services" element={<ServiceListPage />} />
        <Route path="services/new" element={<ServiceCreatePage />} />
        <Route path="services/:id" element={<ServiceDetailPage />} />
        <Route path="services/:id/edit" element={<ServiceEditPage />} />
        <Route
          path="*"
          element={
            <section className="state-panel">
              <p className="eyebrow">404</p>
              <h1>Page not found</h1>
              <p>The requested LaunchRail page does not exist.</p>
            </section>
          }
        />
      </Route>
    </Routes>
  );
}
