from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import threading
import urllib.request

# Worker nodes in our cluster
WORKERS = ["http://localhost:9001"]

def query_worker(node_url, payload, results_list):
    try:
        req_data = json.dumps(payload).encode('utf-8')
        req = urllib.request.Request(
            f"{node_url}/shard/search", 
            data=req_data, 
            headers={'Content-Type': 'application/json'}
        )
        with urllib.request.urlopen(req, timeout=2) as response:
            res = json.loads(response.read().decode('utf-8'))
            results_list.extend(res)
    except Exception as e:
        print(f"Error querying {node_url}: {e}")

class GatewayHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != '/search':
            self.send_response(404)
            self.end_headers()
            return

        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        payload = json.loads(post_data.decode('utf-8'))

        # Concurrent Fan-Out using Threads
        threads = []
        aggregated_results = []
        
        for worker in WORKERS:
            t = threading.Thread(target=query_worker, args=(worker, payload, aggregated_results))
            threads.append(t)
            t.start()

        for t in threads:
            t.join()

        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(aggregated_results).encode('utf-8'))

def run_gateway(port=8080):
    server_address = ('', port)
    httpd = HTTPServer(server_address, GatewayHandler)
    print(f"VectorMesh Gateway running on port {port}...")
    httpd.serve_forever()

if __name__ == '__main__':
    run_gateway()