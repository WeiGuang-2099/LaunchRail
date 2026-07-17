import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import {
  ApiError,
  listServices,
  Service,
  ServiceStatus,
} from "../api/services";

const pageSize = 20;

export function ServiceListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const status = parseStatus(searchParams.get("status"));
  const page = parsePage(searchParams.get("page"));
  const result = searchParams.get("result");
  const [services, setServices] = useState<Service[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const offset = (page - 1) * pageSize;

  useEffect(() => {
    const controller = new AbortController();
    void listServices(status, pageSize, offset, controller.signal)
      .then((response) => {
        setServices(response.items);
        setTotal(response.pagination.total);
        setLoading(false);
      })
      .catch((reason: unknown) => {
        if (reason instanceof DOMException && reason.name === "AbortError") {
          return;
        }
        const requestId =
          reason instanceof ApiError && reason.requestId
            ? ` Request ID: ${reason.requestId}`
            : "";
        setError(`The service catalog could not be loaded.${requestId}`);
        setLoading(false);
      });
    return () => controller.abort();
  }, [offset, retry, status]);

  const lastPage = Math.max(1, Math.ceil(total / pageSize));
  const resultMessage = useMemo(() => {
    if (result === "archived") return "Service archived successfully.";
    return "";
  }, [result]);

  const selectStatus = (nextStatus: ServiceStatus) => {
    setSearchParams({ status: nextStatus, page: "1" });
    setLoading(true);
    setError("");
  };

  const selectPage = (nextPage: number) => {
    setSearchParams({ status, page: String(nextPage) });
    setLoading(true);
    setError("");
  };

  return (
    <>
      <header className="page-header split-header">
        <div>
          <p className="eyebrow">Control plane / Catalog</p>
          <h1>Service catalog</h1>
          <p className="lede">
            A reliable inventory of service ownership, source, and lifecycle.
          </p>
        </div>
        <Link className="button primary" to="/services/new">
          Register service
        </Link>
      </header>

      {resultMessage && (
        <div className="alert success" role="status">
          {resultMessage}
        </div>
      )}

      <section className="catalog-panel" aria-labelledby="catalog-heading">
        <div className="catalog-toolbar">
          <div>
            <h2 id="catalog-heading">Inventory</h2>
            <p>
              {loading
                ? "Loading records"
                : `${total} service${total === 1 ? "" : "s"} in this view`}
            </p>
          </div>
          <div className="status-tabs" aria-label="Service status">
            {(["active", "archived"] as const).map((value) => (
              <button
                key={value}
                className={
                  status === value ? "status-tab active" : "status-tab"
                }
                onClick={() => selectStatus(value)}
                aria-pressed={status === value}
              >
                {value}
              </button>
            ))}
          </div>
        </div>

        {loading && <LoadingList />}
        {!loading && error && (
          <div className="state-panel compact" role="alert">
            <p className="eyebrow danger-text">Catalog unavailable</p>
            <h2>We could not load services</h2>
            <p>{error}</p>
            <button
              className="button secondary"
              onClick={() => {
                setLoading(true);
                setError("");
                setRetry((value) => value + 1);
              }}
            >
              Try again
            </button>
          </div>
        )}
        {!loading && !error && services.length === 0 && (
          <div className="state-panel compact">
            <p className="eyebrow">No records</p>
            <h2>
              {status === "archived"
                ? "No archived services"
                : "Register the first service"}
            </h2>
            <p>
              {status === "archived"
                ? "Archived services will remain visible here for operational history."
                : "Create a durable service record with ownership and repository metadata."}
            </p>
            {status === "active" && (
              <Link className="button primary" to="/services/new">
                Register service
              </Link>
            )}
          </div>
        )}
        {!loading && !error && services.length > 0 && (
          <div className="service-grid">
            {services.map((service) => (
              <ServiceCard key={service.id} service={service} />
            ))}
          </div>
        )}

        {!loading && !error && total > pageSize && (
          <nav className="pagination" aria-label="Catalog pagination">
            <button
              className="button secondary"
              disabled={page <= 1}
              onClick={() => selectPage(page - 1)}
            >
              Previous
            </button>
            <span>
              Page {page} of {lastPage}
            </span>
            <button
              className="button secondary"
              disabled={page >= lastPage}
              onClick={() => selectPage(page + 1)}
            >
              Next
            </button>
          </nav>
        )}
      </section>
    </>
  );
}

function ServiceCard({ service }: { service: Service }) {
  return (
    <article className="service-card">
      <div className="card-topline">
        <span className={`lifecycle ${service.lifecycle}`}>
          {service.lifecycle}
        </span>
        <span className="service-key">{service.serviceKey}</span>
      </div>
      <h3>
        <Link to={`/services/${service.id}`}>{service.displayName}</Link>
      </h3>
      <p className="card-description">
        {service.description || "No service description provided."}
      </p>
      <div className="card-meta">
        <span>Owner</span>
        <strong>{service.ownerTeam}</strong>
      </div>
      <div className="card-footer">
        <span>Updated {formatDate(service.updatedAt)}</span>
        <Link
          to={`/services/${service.id}`}
          aria-label={`View ${service.displayName}`}
        >
          View details
        </Link>
      </div>
    </article>
  );
}

function LoadingList() {
  return (
    <div
      className="service-grid"
      aria-label="Loading services"
      aria-busy="true"
    >
      {[1, 2, 3].map((item) => (
        <div className="service-card skeleton" key={item} aria-hidden="true">
          <span />
          <span />
          <span />
        </div>
      ))}
    </div>
  );
}

function parseStatus(value: string | null): ServiceStatus {
  return value === "archived" || value === "all" ? value : "active";
}

function parsePage(value: string | null): number {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : 1;
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat("en", { dateStyle: "medium" }).format(
    new Date(value),
  );
}
