import React from "react";

function Cars(props) {
  return (
    <div>
      <h1>Car Collection 2024:</h1>
      {props.cars.map((car) => {
        return (
          <div key={car.id}>
            <h2>{car.model}</h2>
            <p>{car.make}</p>
            <p>${car.price}</p>
          </div>
        );
      })}
    </div>
  );
}

export default Cars;
