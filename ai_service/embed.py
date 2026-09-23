import requests
import json

def generate_embedding_and_search(text_query):
    mock_vector = [0.12,0.34,0.56]
    
    payload = {
        "query_vector": mock_vector,
        "top_k": 5
    }
    
    headers = {"Content-Type":"application/json"}
    
    try:
        # Make sure it has /search at the end!
        response = requests.post("http://localhost:8080/search", data=json.dumps(payload), headers=headers)
        print("Gateway Response Status:",response.status_code)
        print("Retrived Shared Results:", response.json())
    except Exception as e:
        print("Error connecting to VectorMesh gateway:", e)
    
if __name__ == "__main__":
    print("Testing VectorMesh Pipeline from Python Client...")
    generate_embedding_and_search("How do I Scale Microservices?")
    