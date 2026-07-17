import { useCallback, useEffect, useState } from "react";
import {
  Link,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import { ApiError, archiveService, getService, Service } from "../api/services";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function ServiceDetailPage() {
  const { id = "" } = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [service, setService] = useState<Service | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [confirming, setConfirming] = useState(false);
  const [archiving, setArchiving] = useState(false);
  const [archiveError, setArchiveError] = useState("");

  const load = useCallback(
    (signal?: AbortSignal) => {
      void getService(id, signal)
        .then((response) => {
          setService(response);
          setLoading(false);
        })
        .catch((reason: unknown) => {
          if (reason instanceof DOMException && reason.name === "AbortError")
            return;
          setError(
            reason instanceof ApiError && reason.status === 404
              ? "Service not found."
              : "The service could not be loaded.",
          );
          setLoading(false);
        });
    },
    [id],
  );

  useEffect(() => {
    const controller = new AbortController();
    load(controller.signal);
    return () => controller.abort();
  }, [load]);

  const confirmArchive = async () => {
    setArchiving(true);
    setArchiveError("");
    try {
      await archiveService(id);
      void navigate("/services?status=archived&page=1&result=archived");
    } catch (reason) {
      const requestId =
        reason instanceof ApiError && reason.requestId
          ? ` Request ID: ${reason.requestId}`
          : "";
      setArchiveError(`The service could not be archived.${requestId}`);
      setArchiving(false);
    }
  };

  if (loading) return <DetailLoading />;
  if (error || service === null) {
    return (
      <section className="state-panel">
        <p className="eyebrow danger-text">Catalog lookup failed</p>
        <h1>{error}</h1>
        <div className="inline-actions">
          <button
            className="button secondary"
            onClick={() => {
              setLoading(true);
              setError("");
              load();
            }}
          >
            Try again
          </button>
          <Link className="button text" to="/services">
            Back to catalog
          </Link>
        </div>
      </section>
    );
  }

  const success = searchParams.get("result");
  return (
    <>
      <header className="page-header detail-header">
        <p className="eyebrow">Catalog / Service record</p>
        <Link
          className="back-link"
          to={`/services?status=${service.archivedAt ? "archived" : "active"}`}
        >
          Back to catalog
        </Link>
        <div className="detail-title-row">
          <div>
            <div className="title-badges">
              <span className={`lifecycle ${service.lifecycle}`}>
                {service.lifecycle}
              </span>
              {service.archivedAt && (
                <span className="archived-badge">archived</span>
              )}
            </div>
            <h1>{service.displayName}</h1>
            <p className="service-key large">{service.serviceKey}</p>
          </div>
          {!service.archivedAt && (
            <div className="inline-actions">
              <Link
                className="button secondary"
                to={`/services/${service.id}/edit`}
              >
                Edit metadata
              </Link>
              <button
                className="button danger-outline"
                onClick={() => setConfirming(true)}
              >
                Archive service
              </button>
            </div>
          )}
        </div>
      </header>

      {(success === "created" || success === "updated") && (
        <div className="alert success" role="status">
          Service {success === "created" ? "registered" : "updated"}{" "}
          successfully.
        </div>
      )}
      {archiveError && <div className="alert error">{archiveError}</div>}

      <section className="detail-layout">
        <article className="detail-card primary-detail">
          <p className="eyebrow">Service context</p>
          <h2>About this service</h2>
          <p className="description-copy">
            {service.description || "No service description provided."}
          </p>
          <dl className="metadata-grid">
            <div>
              <dt>Owner team</dt>
              <dd>{service.ownerTeam}</dd>
            </div>
            <div>
              <dt>Lifecycle</dt>
              <dd>{service.lifecycle}</dd>
            </div>
            <div className="wide">
              <dt>Source repository</dt>
              <dd>
                <a
                  href={service.repositoryUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  {service.repositoryUrl}
                </a>
              </dd>
            </div>
          </dl>
        </article>
        <aside className="detail-card timeline-card">
          <p className="eyebrow">Record history</p>
          <dl>
            <div>
              <dt>Created</dt>
              <dd>{formatTimestamp(service.createdAt)}</dd>
            </div>
            <div>
              <dt>Last updated</dt>
              <dd>{formatTimestamp(service.updatedAt)}</dd>
            </div>
            {service.archivedAt && (
              <div>
                <dt>Archived</dt>
                <dd>{formatTimestamp(service.archivedAt)}</dd>
              </div>
            )}
          </dl>
        </aside>
      </section>

      {confirming && (
        <ConfirmDialog
          title={`Archive ${service.displayName}?`}
          description="The record will remain available for operational history, but its metadata can no longer be edited."
          confirmLabel="Archive service"
          busy={archiving}
          onCancel={() => setConfirming(false)}
          onConfirm={() => void confirmArchive()}
        />
      )}
    </>
  );
}

function DetailLoading() {
  return (
    <section className="state-panel" aria-busy="true">
      <p className="eyebrow">Loading service</p>
      <h1>Retrieving catalog record</h1>
    </section>
  );
}

function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
