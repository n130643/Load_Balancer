# Load_Balancer
Load balancer app serves two APIs for balacing the load on system from incoming requests: 
    1. Register API
    2. Proxy API

**Register API** : 
  A POST API servers to register a route (url) in load balancing sytem that later will get incoming requests routed from load balancer. It takes url         through request body. Empty body results in error response.

  Success Reponse :
  
  <img width="550" alt="Screenshot 2022-07-04 at 2 53 21 PM" src="https://user-images.githubusercontent.com/32019167/177125154-f92d7386-af92-4eee-9578-7c2f8fc83e9f.png">

  Error Response:
  
  <img width="550" alt="Screenshot 2022-07-04 at 2 51 55 PM" src="https://user-images.githubusercontent.com/32019167/177124938-b57623c9-3945-4ab0-a447-b56e032d7448.png">


  
 **Proxy API** :
  An API that works with all http verbs servers to distribute incoming requests across registered routes in load balancing system.
  
  Success Reponse :

  <img width="550" alt="Screenshot 2022-07-04 at 2 56 03 PM" src="https://user-images.githubusercontent.com/32019167/177125656-f98ab375-67c1-4d81-9587-5d542c2ca769.png">

  
  Error Response :
  
<img width="550" alt="Screenshot 2022-07-04 at 2 49 31 PM" src="https://user-images.githubusercontent.com/32019167/177124707-ff653d9e-1673-42ff-a0c7-1f45664f2ebc.png">

  
      
