import requests
import json

def test_rag_pipeline():
    payload = {
        "query_text": "How does VectorMesh handle scaling?",
        "query_vector": [0.1, 0.2, 0.3],
        "top_k": 2
    }
    
    headers = {"Content-Type": "application/json"}
    
    try:
        response = requests.post("http://localhost:8080/rag/search", data=json.dumps(payload), headers=headers)
        print("Go Gateway Response Status:", response.status_code)
        data = response.json()
        
        # Print cache status here!
        print(f"⚡ Served from LRU Cache: {data.get('cached')}")
        
        print("\n--- Retrieved RAG Context Chunks ---")
        for chunk in data.get("context_chunks", []):
            print(f"[{chunk['id']}] (Score: {chunk['score']:.2f}): {chunk['text']}")
            
        print("\n--- Final Assembled LLM Prompt ---")
        print(data.get("generated_prompt"))
    except Exception as e:
        print("Error connecting to Go Gateway:", e)

if __name__ == "__main__":
    print("Testing Go-based VectorMesh RAG Pipeline...")
    test_rag_pipeline()