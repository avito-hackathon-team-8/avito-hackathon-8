import { RouterProvider } from "react-router";

import { browserRouter } from "@/app/routes";

function App() {
  return <RouterProvider router={browserRouter} />;
}

export default App;
