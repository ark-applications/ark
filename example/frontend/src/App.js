import axios from "axios";
import { useEffect, useState } from "react";

import "./App.css";
import Cars from "./cars";

const API_URL = "http://localhost:3000/api/v1/cars.json";

const getApiData = async () => {
  const resp = await axios.get(API_URL);
  return resp.data;
};

function App() {
  const [cars, setCars] = useState([]);

  useEffect(() => {
    let mounted = true;
    getApiData().then((cars) => {
      if (mounted) {
        setCars(cars);
      }
    });
    return () => (mounted = false);
  }, []);

  return (
    <div className="App">
      <Cars cars={cars} />
    </div>
  );
}

export default App;
