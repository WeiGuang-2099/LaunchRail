import type { ServiceFormValues } from "./ServiceForm";

export const emptyServiceForm: ServiceFormValues = {
  serviceKey: "",
  displayName: "",
  description: "",
  ownerTeam: "",
  repositoryUrl: "",
  lifecycle: "development",
};
