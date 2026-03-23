import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 10000,
  duration: '20s',
};

function gerarSensor() {
  const sensores = ["Potenciométrico", "Termistor", "Ultrassônico", "LDR", "Infravermelho", "Temperatura", "Pressão", "Umidade", "Movimento"];
  const leituras = ["analogico", "discreto"];

  const sensor = sensores[Math.floor(Math.random() * sensores.length)];
  const leitura = leituras[Math.floor(Math.random() * leituras.length)];

  return {
    id: Math.floor(Math.random() * 1000000).toString(),
    timestamp: new Date().toISOString(),
    "tipo-sensor": sensor,
    "tipo-leitura": leitura,
    valor: leitura === "digital"
      ? Math.round(Math.random()) // 0 ou 1
      : Number((Math.random() * 5000).toFixed(2))
  };
}

export default function () {
  // Make a GET request to the target URL
  http.post('http://192.168.96.1:8080/dados-sensores', JSON.stringify(gerarSensor()), {'Content-Type': 'application/json'});

  // Sleep for 1 second to simulate real-world usage
  sleep(1);
}