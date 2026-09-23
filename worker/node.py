from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import math

def cosine_similarity(a, b):
    if len(a) != len(b) or len(a) == 0:
        return 0.0
    dot_product = sum(x * y for x, y in zip(a, b))
    norm_a = math.sqrt(sum(x * x for x in a))
    norm_b = math.sqrt(sum(y * y for y in b))
    if norm_a == 0 or norm_b == 0:
        return 0.0
    return dot_product / (norm_a * norm_b)

class ShardHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        req = json.loads(post_data.decode('utf-8'))
        
        query_vector = req.get("query_vector", [0.1, 0.2, 0.3])
        
        # Shard dataset containing vectors AND text chunks for RAG
        shard_data = [
            {"id": "doc_101", "text": "VectorMesh uses concurrent Go routines to fan-out queries across distributed database shards.", "vector": [0.1, 0.2, 0.3]},
            {"id": "doc_102", "text": "RAG pipelines combine semantic vector search with LLM context window assembly.", "vector": [0.4, 0.5, 0.6]}
        ]
        
        results = []
        for doc in shard_data:
            score = cosine_similarity(query_vector, doc["vector"])
            results.append({
                "id": doc["id"],
                "text": doc["text"],
                "score": score
            })
            
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(results).encode('utf-8'))

def run_worker(port=9001):
    server_address = ('', port)
    httpd = HTTPServer(server_address, ShardHandler)
    print(f"VectorMesh RAG Shard Worker running on port {port}...")
    httpd.serve_forever()

if __name__ == '__main__':
    run_worker()