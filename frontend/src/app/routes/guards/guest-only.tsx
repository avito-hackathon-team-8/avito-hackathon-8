import { Navigate, Outlet } from "react-router";

import { useCurrentUser } from "@/entities/user";
import { APP_ROUTES } from "@/shared/config/routes";

export const GuestOnly = () => {
  const { data: currentUser, isLoading } = useCurrentUser();

  if (isLoading) {
    return null;
  }

  if (currentUser) {
    return <Navigate to={APP_ROUTES.home} replace />;
  }

  return <Outlet />;
};
