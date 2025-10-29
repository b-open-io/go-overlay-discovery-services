// Package slap implements the SLAP (Service Lookup Availability Protocol) lookup service functionality.
// This file contains the documentation for the SLAP lookup service, separated from the implementation
// to improve code organization and maintainability.
package slap

// LookupDocumentation contains the comprehensive documentation for the SLAP lookup service.
// This documentation describes how to use the service, including query formats, examples,
// and important considerations for developers.
const LookupDocumentation = `# SLAP Lookup Service

**Protocol Name**: SLAP (Service Lookup Availability Protocol)
**Lookup Service Name**: ` + "`SLAPLookupService`" + `

---

## Overview

The SLAP Lookup Service is used to **query** the known SLAP tokens in your overlay database. It allows you to discover nodes that have published SLAP outputs, indicating they offer specific services (prefixed ` + "`ls_`" + `).

This lookup service is typically invoked by sending a [LookupQuestion](https://www.npmjs.com/package/@bsv/overlay#lookupservice) with:
- ` + "`question.service = 'ls_slap'`" + `
- ` + "`question.query`" + ` containing parameters for searching.

---

## Purpose

- **Discovery**: Find all services that have been advertised with the SLAP protocol.
- **Filtering**: Narrow results by domain or by the ` + "`ls_`" + ` service name.

---

## Querying the SLAP Lookup Service

When you call ` + "`lookup(question)`" + ` on the SLAP Lookup Service, you must include:

1. **` + "`question.service`" + `** set to ` + "`\"ls_slap\"`" + `.
2. **` + "`question.query`" + `**: Can be one of the following:
   - ` + "`\"findAll\"`" + ` (string literal): Returns **all** known SLAP records.
   - An object of type:
     ` + "```" + `ts
     interface SLAPQuery {
       domain?: string
       service?: string
     }
     ` + "```" + `
     where:
     - ` + "`domain`" + ` is an optional string. If provided, results will match that domain/advertisedURI.
     - ` + "`service`" + ` is an optional string. If provided, results will match services with that name (typically prefixed ` + "`ls_`" + `).

### Examples

1. **Find all SLAP records**:
   ` + "```" + `go
   import "github.com/bsv-blockchain/go-sdk/overlay/lookup"

   resolver := lookup.NewLookupResolver()
   results, err := resolver.Query(ctx, &lookup.LookupQuestion{
       Service: "ls_slap",
       Query:   "findAll",
   }, 10000)
   ` + "```" + `

2. **Find by domain**:
   ` + "```" + `go
   results, err := resolver.Query(ctx, &lookup.LookupQuestion{
       Service: "ls_slap",
       Query: map[string]interface{}{
           "domain": "https://mylookup.example",
       },
   }, 10000)
   ` + "```" + `

3. **Find by service (most common)**:
   ` + "```" + `go
   results, err := resolver.Query(ctx, &lookup.LookupQuestion{
       Service: "ls_slap",
       Query: map[string]interface{}{
           "service": "ls_treasury",
       },
   }, 10000)
   ` + "```" + `

4. **Find by domain AND service**:
   ` + "```" + `go
   results, err := resolver.Query(ctx, &lookup.LookupQuestion{
       Service: "ls_slap",
       Query: map[string]interface{}{
           "domain":  "https://mylookup.example",
           "service": "ls_treasury",
       },
   }, 10000)
   ` + "```" + `

---

## Gotchas and Tips

- **Service Prefix**: The SLAP manager expects services to start with ` + "`ls_`" + `. If you see no results, ensure you used the correct prefix.
- **Strict Matching**: Domain matching requires an exact string match. If you have a different protocol (https vs https+bsvauth vs https+bsvauth+smf), be sure to store/lookup accordingly.
- **Partial Queries**: If you only provide ` + "`service`" + `, domain-based filtering is not applied, and vice versa.
- **Single Service**: Unlike SHIP's topics array, SLAP queries filter by a single service name.

---

## Further Reading

- **SLAPTopicManager**: For how the outputs are admitted.
- **BRC-101 Overlays**: The general pattern for these sorts of services.
- **SHIP**: The complementary protocol for topic hosting advertisements.
`
