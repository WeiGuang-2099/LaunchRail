import { NavLink, Outlet } from "react-router-dom";

export function AppShell() {
  return (
    <div className="app-shell">
      <aside className="rail-sidebar">
        <NavLink className="brand" to="/services" aria-label="LaunchRail home">
          <span className="brand-code">LR / 01</span>
          <span className="wordmark">
            Launch
            <br />
            Rail
          </span>
        </NavLink>

        <div className="rail-line" aria-hidden="true">
          <span className="rail-stop active" />
          <span className="rail-stop" />
          <span className="rail-stop" />
        </div>

        <div className="sidebar-section">
          <p>Control plane</p>
          <nav aria-label="Primary navigation">
            <NavLink
              to="/services"
              className={({ isActive }) =>
                isActive ? "nav-link active" : "nav-link"
              }
            >
              <span>01</span>
              Service manifest
            </NavLink>
          </nav>
        </div>

        <div className="sidebar-status">
          <span className="status-light" aria-hidden="true" />
          <span>
            <small>Environment</small>
            Local / connected
          </span>
        </div>
      </aside>

      <div className="workspace">
        <header className="workspace-bar">
          <span>LaunchRail / Internal developer platform</span>
          <div>
            <span>Registry</span>
            <strong>Services</strong>
          </div>
        </header>

        <main className="page-frame">
          <Outlet />
        </main>

        <footer className="footer">
          <span>LaunchRail control plane</span>
          <span>Local registry / v0.1</span>
        </footer>
      </div>
    </div>
  );
}
