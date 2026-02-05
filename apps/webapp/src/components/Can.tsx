import { ParentComponent } from "solid-js";
import { useUserContext } from "../providers/user";

interface CanProps {
  permission: string;
  fallback?: ParentComponent;
}

export const Can: ParentComponent<CanProps> = (props) => {
  const userContext = useUserContext();

  return (
    <>
      {userContext.can(props.permission) ? (
        props.children
      ) : props.fallback ? (
        props.fallback(props)
      ) : null}
    </>
  );
};
