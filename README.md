# Rate Limiter

Rate Limiter is a simple implementation of two rate-limiting strategies:

1. **Token Fixed Window (Per User)**
   - Each user is assigned a fixed number of tokens.
   - Tokens expire after a specified time window.
   - If the number of tokens consumed by a user exceeds a limit within the window, the request is not served.

2. **Shared Bucket Token**
   - Uses a Redis Sorted Set to manage tokens.
   - A background goroutine adds tokens to the shared bucket if the limit is not reached.
   - Requests are served if there are available tokens in the bucket.

3. **Sliding Window**
  - Utilize a Redis sorted set to store timestamps(as score) of requests.
   - Periodically check the set to analyze the timestamps within the desired time window (e.g., the last N seconds).
   - Adjust the set by removing timestamps outside the time window.
   - Calculate the count of remaining timestamps within the window to track the rate of requests.
