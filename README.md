# Protocol Rewards

config.hjson
```hjson
{
   providers: [
      https://atlasnet.rpc.mavryk.network/
   ]
   mvkt_providers: [
       https://atlasnet.api.mavryk.network/
   ]
   database: {
      host: 127.0.0.1
      port: 5432
      user: protocol_rewards
      password: protocol_rewards
      database: protocol_rewards
   }
   storage: {
      mode: rolling
      stored_cycles: 20
   }
   discord_notificator: {
      webhook_url: url
      webhook_id: id
      webhook_token: token
   }
   // optional subset if wanted, if not just delete it or keep it empty
   delegates: [
      mv1VNRtHZdLzSJfyvvz2cxAoR1kWoNDWMisL
      mv1CtCq3D2RrCxx6VL5aMTkfq8tYLSK4sXmN
      mv1SZS8SZB5Wt5GTMnLxqxn13pAXrSNMsXQ1
   ]
}
```

.env
```
LOG_LEVEL=debug
LISTEN=127.0.0.1:3000
PRIVATE_LISTEN=127.0.0.1:4000

```

LOG_LEVEL accepted values are debug, info, warn, error. Defaults to info level.

U can define env variables in the .env file or in your environment directly as you choose. If you forgot to define your env variable they will be assigned the default values.

testing command flags
```
go run main.go -log debug -test mv1VW2QKBXfsroTFkdaS5xejZXbmpGrxYu6u:745
```

### Credits

**Powered by [MvKT API](https://atlasnet.api.mavryk.network/)** - `protocol-rewards` use MVKT api to fetch unstake requests.