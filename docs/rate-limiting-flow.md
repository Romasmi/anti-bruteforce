# Rate Limiting Flow

How the service processes a `CheckAuth` request.

```mermaid
sequenceDiagram
    participant GW as API Gateway
    participant AB as Anti-Bruteforce
    participant DB as PostgreSQL<br/>(white/black lists)
    participant BK as Leaky Bucket<br/>(Redis or in-memory)

    GW->>AB: CheckAuth(login, password, ip)
    AB->>AB: Validate input

    note over AB,BK: Login, password, and IP strategies run in parallel

    par Login bucket
        AB->>BK: Allow? key=login
        BK-->>AB: allowed / denied
    and Password bucket
        AB->>BK: Allow? key=password
        BK-->>AB: allowed / denied
    and IP checks
        AB->>DB: ip in whitelist?
        DB-->>AB: yes / no
        alt whitelisted
            note over AB: IP strategy → allow, skip bucket
        else not whitelisted
            AB->>DB: ip in blacklist?
            DB-->>AB: yes / no
            alt blacklisted
                note over AB: IP strategy → deny, skip bucket
            else not blacklisted
                AB->>BK: Allow? key=ip
                BK-->>AB: allowed / denied
            end
        end
    end

    alt all three strategies allowed
        AB-->>GW: ok=true
    else any strategy denied
        AB-->>GW: ok=false
    end
```

## Bucket behaviour

Each bucket is a leaky bucket: it drains at a constant rate and is capped at a maximum level.
A request is allowed when adding one token does not exceed the capacity; denied otherwise.

| Strategy | Default capacity | Window |
|---|---|---|
| Login | 10 | 60 s |
| Password | 1000 | 60 s |
| IP | 1000 | 60 s |

Capacities are configurable via `configs/config.yaml`.

## Storage

When Redis is configured the buckets are persisted there and shared across instances.
Without Redis each instance keeps its own in-memory buckets.

White/black lists are always stored in PostgreSQL and consulted on every request.
