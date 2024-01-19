# Rate Limiter

Rate Limiter is a simple implementation of three rate-limiting strategies:

1. **Token Fixed Window (Per User)**

   
   ![fixed_window](https://github.com/tiffany831101/rate-limiter/assets/39373272/0f5a42a2-2b9c-4bf0-9aa5-81fa60cd2698)

   - Each user is assigned a fixed number of tokens.
   - Tokens expire after a specified time window.
   - If the number of tokens consumed by a user exceeds a limit within the window, the request is not served.

3. **Shared Bucket Token**
        
     ![bucket_token](https://github.com/tiffany831101/rate-limiter/assets/39373272/998ce8e5-f0e9-4294-930f-95e9004b6239)
   
   - Uses a Redis Sorted Set to manage tokens.
   - A background goroutine adds tokens to the shared bucket if the limit is not reached.
   - Requests are served if there are available tokens in the bucket.
  

4. **Sliding Window**
   - Utilize a Redis sorted set to store timestamps(as score) of requests.
   - Periodically check the set to analyze the timestamps within the desired time window (e.g., the last N seconds).
   - Adjust the set by removing timestamps outside the time window.
   - Calculate the count of remaining timestamps within the window to track the rate of requests.
