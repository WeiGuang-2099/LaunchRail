import { NavLink, Outlet } from "react-router-dom";

export function AppShell() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <NavLink className="brand" to="/services" aria-label="LaunchRail home">
          <span className="brand-mark" aria-hidden="true">
            LR
          </span>
          <span>
            <strong>LaunchRail</strong>
            <small>Internal Developer Platform</small>
          </span>
        </NavLink>
        <nav aria-label="Primary navigation">
          <NavLink
            to="/services"
            className={({ isActive }) =>
              isActive ? "nav-link active" : "nav-link"
            }
          >
            Service catalog
          </NavLink>
        </nav>
        <div className="environment-badge">Local environment</div>
      </header>
      <main className="page-frame">
        <Outlet />
      </main>
      <footer className="footer">
        <span>LaunchRail control plane</span>
        <span>Service catalog milestone</span>
      </footer>
    </div>
  );
}
