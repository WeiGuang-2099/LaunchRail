import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getService, Service, updateService } from "../api/services";
import { ServiceForm } from "../components/ServiceForm";

export function ServiceEditPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [service, setService] = useState<Service | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void getService(id, controller.signal)
      .then((response) => {
        setService(response);
        setLoading(false);
      })
      .catch((reason: unknown) => {
        if (reason instanceof DOMException && reason.name === "AbortError")
          return;
        setError("The service could not be loaded for editing.");
        setLoading(false);
      });
    return () => controller.abort();
  }, [id]);

  if (loading) {
    return (
      <section className="state-panel" aria-busy="true">
        <h1>Loading service</h1>
      </section>
    );
  }
  if (error || service === null) {
    return (
      <section className="state-panel">
        <h1>{error}</h1>
        <Link to="/services">Back to catalog</Link>
      </section>
    );
  }
  if (service.archivedAt) {
    return (
      <section className="state-panel">
        <p className="eyebrow danger-text">Archived record</p>
        <h1>This service cannot be edited</h1>
        <p>Archived services are retained as immutable operational history.</p>
        <Link className="button secondary" to={`/services/${service.id}`}>
          View service
        </Link>
      </section>
    );
  }

  return (
    <>
      <header className="page-header">
        <p className="eyebrow">Manifest / Amend record</p>
        <Link className="back-link" to={`/services/${service.id}`}>
          Back to service
        </Link>
        <h1>Edit {service.displayName}</h1>
        <p className="lede">
          The service key is permanent; all other catalog metadata can evolve.
        </p>
      </header>
      <ServiceForm
        initialValues={{
          serviceKey: service.serviceKey,
          displayName: service.displayName,
          description: service.description,
          ownerTeam: service.ownerTeam,
          repositoryUrl: service.repositoryUrl,
          lifecycle: service.lifecycle,
        }}
        serviceKeyEditable={false}
        submitLabel="Save changes"
        onSubmit={async (values) => {
          const updated = await updateService(service.id, {
            displayName: values.displayName,
            description: values.description,
            ownerTeam: values.ownerTeam,
            repositoryUrl: values.repositoryUrl,
            lifecycle: values.lifecycle,
          });
          void navigate(`/services/${updated.id}?result=updated`);
        }}
      />
    </>
  );
}
