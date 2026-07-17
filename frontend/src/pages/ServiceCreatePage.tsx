import { Link, useNavigate } from "react-router-dom";
import { createService } from "../api/services";
import { ServiceForm } from "../components/ServiceForm";
import { emptyServiceForm } from "../components/serviceFormDefaults";

export function ServiceCreatePage() {
  const navigate = useNavigate();
  return (
    <>
      <header className="page-header">
        <p className="eyebrow">Catalog / Register</p>
        <Link className="back-link" to="/services">
          Back to catalog
        </Link>
        <h1>Register a service</h1>
        <p className="lede">
          Create the canonical record that platform workflows will build on.
        </p>
      </header>
      <ServiceForm
        initialValues={emptyServiceForm}
        serviceKeyEditable
        submitLabel="Register service"
        onSubmit={async (values) => {
          const service = await createService(values);
          void navigate(`/services/${service.id}?result=created`);
        }}
      />
    </>
  );
}
