

import http from 'k6/http';
import { sleep, check } from 'k6';
export const options = {
	stages: [
		{ duration: '1m', target: 200 }, // ramp up
	
	],
};



export default () => {
    const reqBody = {
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636, 
    "dst_lat": -7.581681866327657, 
    "dst_lon": 110.82648949172574
    }

    const res = http.post("http://localhost:3000/api/navigations/shortestPath", JSON.stringify(reqBody), {
         headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json',
      }
    });
  check(res, { '200': (r) => r.status === 200 });
    sleep(1);

}