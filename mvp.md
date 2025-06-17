# MVP CHECKLIST

| Feature                   | Tools                                                 |
| ------------------------- | ----------------------------------------------------- |
| Create Users              | create freelancers and clients and admin              |
| Generate wallets          | `keypair.Random()`                                    |
| Create invoice            | generate invoice data and send to the client          |
| Track payments            | In house and custom                                   |
| Send payments             | `txnbuild.Payment + Sign + Submit`                    |
| USDC handling             | Use Circle issuer: `USDC:G...`                        |
| Match payments to invoice | Use memo fields                                       |
| Bridge support            | Use external tool like Allbridge (via web or webhook) |
