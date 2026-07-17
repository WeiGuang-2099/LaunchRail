import { FormEvent, useState } from "react";
import { ApiError, Lifecycle, ServiceCreateInput } from "../api/services";

export type ServiceFormValues = ServiceCreateInput;

interface ServiceFormProps {
  initialValues: ServiceFormValues;
  serviceKeyEditable: boolean;
  submitLabel: string;
  onSubmit: (values: ServiceFormValues) => Promise<void>;
}

type FieldErrors = Partial<Record<keyof ServiceFormValues, string>>;

export function ServiceForm({
  initialValues,
  serviceKeyEditable,
  submitLabel,
  onSubmit,
}: ServiceFormProps) {
  const [values, setValues] = useState(initialValues);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const setValue = <K extends keyof ServiceFormValues>(
    key: K,
    value: ServiceFormValues[K],
  ) => {
    setValues((current) => ({ ...current, [key]: value }));
    setFieldErrors((current) => ({ ...current, [key]: undefined }));
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const validationErrors = validate(values, serviceKeyEditable);
    setFieldErrors(validationErrors);
    setFormError("");
    if (Object.keys(validationErrors).length > 0) {
      return;
    }

    setSubmitting(true);
    try {
      await onSubmit(values);
    } catch (error) {
      if (error instanceof ApiError) {
        const serverErrors: FieldErrors = {};
        for (const detail of error.details) {
          if (isFormField(detail.field)) {
            serverErrors[detail.field] = detail.message;
          }
        }
        setFieldErrors(serverErrors);
        setFormError(
          Object.keys(serverErrors).length === 0
            ? `${error.message}${error.requestId ? ` Request ID: ${error.requestId}` : ""}`
            : "Review the highlighted fields and try again.",
        );
      } else {
        setFormError(
          "The service could not be saved. Check your connection and try again.",
        );
      }
      setSubmitting(false);
    }
  };

  return (
    <form
      className="service-form"
      onSubmit={(event) => void handleSubmit(event)}
      noValidate
    >
      {formError && (
        <div className="alert error" role="alert">
          {formError}
        </div>
      )}

      <div className="form-section">
        <div className="section-heading">
          <span>01</span>
          <div>
            <h2>Identity</h2>
            <p>Stable identifiers used by platform automation and people.</p>
          </div>
        </div>
        <div className="form-grid">
          <Field
            label="Service key"
            error={fieldErrors.serviceKey}
            hint="Lowercase letters, numbers, and hyphens."
          >
            <input
              name="serviceKey"
              value={values.serviceKey}
              onChange={(event) => setValue("serviceKey", event.target.value)}
              disabled={!serviceKeyEditable}
              aria-invalid={Boolean(fieldErrors.serviceKey)}
              aria-describedby="serviceKey-help serviceKey-error"
            />
          </Field>
          <Field label="Display name" error={fieldErrors.displayName}>
            <input
              name="displayName"
              value={values.displayName}
              onChange={(event) => setValue("displayName", event.target.value)}
              aria-invalid={Boolean(fieldErrors.displayName)}
              aria-describedby="displayName-error"
            />
          </Field>
        </div>
      </div>

      <div className="form-section">
        <div className="section-heading">
          <span>02</span>
          <div>
            <h2>Ownership and source</h2>
            <p>
              Connect the service to its accountable team and source repository.
            </p>
          </div>
        </div>
        <div className="form-grid">
          <Field label="Owner team" error={fieldErrors.ownerTeam}>
            <input
              name="ownerTeam"
              value={values.ownerTeam}
              onChange={(event) => setValue("ownerTeam", event.target.value)}
              aria-invalid={Boolean(fieldErrors.ownerTeam)}
              aria-describedby="ownerTeam-error"
            />
          </Field>
          <Field
            label="GitHub repository"
            error={fieldErrors.repositoryUrl}
            hint="Use an HTTPS github.com owner/repository URL."
          >
            <input
              name="repositoryUrl"
              type="url"
              value={values.repositoryUrl}
              onChange={(event) =>
                setValue("repositoryUrl", event.target.value)
              }
              aria-invalid={Boolean(fieldErrors.repositoryUrl)}
              aria-describedby="repositoryUrl-help repositoryUrl-error"
            />
          </Field>
        </div>
      </div>

      <div className="form-section">
        <div className="section-heading">
          <span>03</span>
          <div>
            <h2>Operational context</h2>
            <p>Describe the service and its current delivery lifecycle.</p>
          </div>
        </div>
        <div className="form-grid">
          <Field label="Lifecycle" error={fieldErrors.lifecycle}>
            <select
              name="lifecycle"
              value={values.lifecycle}
              onChange={(event) =>
                setValue("lifecycle", event.target.value as Lifecycle)
              }
              aria-invalid={Boolean(fieldErrors.lifecycle)}
            >
              <option value="development">Development</option>
              <option value="production">Production</option>
              <option value="deprecated">Deprecated</option>
            </select>
          </Field>
          <Field
            label="Description"
            error={fieldErrors.description}
            hint={`${values.description.length}/2000 characters`}
            wide
          >
            <textarea
              name="description"
              rows={5}
              value={values.description}
              onChange={(event) => setValue("description", event.target.value)}
              aria-invalid={Boolean(fieldErrors.description)}
              aria-describedby="description-help description-error"
            />
          </Field>
        </div>
      </div>

      <div className="form-actions">
        <button className="button primary" type="submit" disabled={submitting}>
          {submitting ? "Saving..." : submitLabel}
        </button>
      </div>
    </form>
  );
}

interface FieldProps {
  label: string;
  error?: string;
  hint?: string;
  wide?: boolean;
  children: React.ReactNode;
}

function Field({ label, error, hint, wide, children }: FieldProps) {
  const id = label
    .toLowerCase()
    .replaceAll(" ", "")
    .replace("githubrepository", "repositoryUrl");
  const fieldId =
    id === "servicekey"
      ? "serviceKey"
      : id === "displayname"
        ? "displayName"
        : id === "ownerteam"
          ? "ownerTeam"
          : id;
  return (
    <label className={wide ? "field wide" : "field"}>
      <span>{label}</span>
      {children}
      {hint && <small id={`${fieldId}-help`}>{hint}</small>}
      {error && (
        <small className="field-error" id={`${fieldId}-error`} role="alert">
          {error}
        </small>
      )}
    </label>
  );
}

function validate(
  values: ServiceFormValues,
  validateKey: boolean,
): FieldErrors {
  const errors: FieldErrors = {};
  if (validateKey && !/^[a-z][a-z0-9-]{0,62}$/.test(values.serviceKey.trim())) {
    errors.serviceKey =
      "Use 1-63 lowercase letters, numbers, or hyphens, starting with a letter.";
  }
  requiredLength(errors, "displayName", values.displayName, 120);
  requiredLength(errors, "ownerTeam", values.ownerTeam, 120);
  if (values.description.length > 2000) {
    errors.description = "Description must be at most 2000 characters.";
  }
  if (!isGitHubRepository(values.repositoryUrl)) {
    errors.repositoryUrl = "Enter an HTTPS GitHub owner and repository URL.";
  }
  return errors;
}

function requiredLength(
  errors: FieldErrors,
  field: "displayName" | "ownerTeam",
  value: string,
  maximum: number,
) {
  const length = value.trim().length;
  if (length === 0) {
    errors[field] = "This field is required.";
  } else if (length > maximum) {
    errors[field] = `Must be at most ${maximum} characters.`;
  }
}

function isGitHubRepository(value: string): boolean {
  try {
    const url = new URL(value);
    const parts = url.pathname.replace(/^\/+|\/+$/g, "").split("/");
    return (
      url.protocol === "https:" &&
      url.hostname.toLowerCase() === "github.com" &&
      url.port === "" &&
      url.username === "" &&
      url.password === "" &&
      url.search === "" &&
      url.hash === "" &&
      parts.length === 2 &&
      parts.every(Boolean)
    );
  } catch {
    return false;
  }
}

function isFormField(value: string): value is keyof ServiceFormValues {
  return [
    "serviceKey",
    "displayName",
    "description",
    "ownerTeam",
    "repositoryUrl",
    "lifecycle",
  ].includes(value);
}
