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
      <header className="page-header manifest-header">
        <div>
          <p className="eyebrow">Registry 01 / Service records</p>
          <h1>
            Service
            <span>manifest</span>
          </h1>
        </div>
        <div className="header-brief">
          <p className="lede">
            The plain record of what runs, who owns it, and where the code
            lives.
          </p>
          <Link className="button primary" to="/services/new">
            Register service
          </Link>
        </div>
      </header>

      {resultMessage && (
        <div className="alert success" role="status">
          <span>Operation complete</span>
          {resultMessage}
        </div>
      )}

      <section className="manifest" aria-labelledby="catalog-heading">
        <div className="manifest-toolbar">
          <div className="manifest-title">
            <span>Live index</span>
            <h2 id="catalog-heading">Registered services</h2>
          </div>
          <div className="status-tabs" aria-label="Service status">
            {(["active", "archived"] as const).map((value, index) => (
              <button
                key={value}
                className={
                  status === value ? "status-tab active" : "status-tab"
                }
                onClick={() => selectStatus(value)}
                aria-pressed={status === value}
              >
                <span>{String(index + 1).padStart(2, "0")}</span>
                {value}
              </button>
            ))}
          </div>
          <p className="record-count" aria-live="polite">
            <strong>{loading ? "--" : String(total).padStart(2, "0")}</strong>
            records in view
          </p>
        </div>

        <div className="manifest-columns" aria-hidden="true">
          <span>No.</span>
          <span>Service / key</span>
          <span>Owner</span>
          <span>Lifecycle</span>
          <span>Updated</span>
          <span>Record</span>
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
            <p className="eyebrow">No records on this line</p>
            <h2>
              {status === "archived"
                ? "No archived services"
                : "Register the first service"}
            </h2>
            <p>
              {status === "archived"
                ? "Archived services stay here as part of the operational record."
                : "Start the manifest with an owner, repository, and lifecycle."}
            </p>
            {status === "active" && (
              <Link className="button primary" to="/services/new">
                Register service
              </Link>
            )}
          </div>
        )}
        {!loading && !error && services.length > 0 && (
          <div className="manifest-list">
            {services.map((service, index) => (
              <ServiceRow
                key={service.id}
                service={service}
                index={offset + index + 1}
              />
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
              Page {page} / {lastPage}
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

function ServiceRow({ service, index }: { service: Service; index: number }) {
  return (
    <Link
      className="service-row"
      to={`/services/${service.id}`}
      aria-label={`View ${service.displayName}`}
    >
      <span className="row-number">{String(index).padStart(2, "0")}</span>
      <span className="service-identity">
        <strong>{service.displayName}</strong>
        <small>{service.serviceKey}</small>
      </span>
      <span className="row-owner" data-label="Owner">
        {service.ownerTeam}
      </span>
      <span className={`lifecycle ${service.lifecycle}`}>
        {service.lifecycle}
      </span>
      <span className="row-updated" data-label="Updated">
        {formatDate(service.updatedAt)}
      </span>
      <span className="row-action">View -&gt;</span>
    </Link>
  );
}

function LoadingList() {
  return (
    <div
      className="manifest-list"
      aria-label="Loading services"
      aria-busy="true"
    >
      {[1, 2, 3].map((item) => (
        <div className="service-row skeleton-row" key={item} aria-hidden="true">
          <span />
          <span />
          <span />
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
