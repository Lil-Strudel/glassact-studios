# GlassAct Studios - Domain Rules

These rules define the business logic and domain constraints for the GlassAct Studios platform.

## Business Overview

GlassAct Studios manufactures custom stained glass inlays for gravestones. The platform serves B2B customers (gravestone engravers called "dealerships") who order inlays on behalf of end consumers.

### Key Stakeholders

| Stakeholder | Role |
|-------------|------|
| Dealership | Orders inlays, approves designs, pays invoices |
| GlassAct Designer | Creates proofs, responds to design feedback |
| GlassAct Production | Manages manufacturing workflow |
| GlassAct Billing | Creates and manages invoices |

## Entity Lifecycles

### Project Status Flow

```
                    ┌─────────────────────────────────────────────┐
                    │                                             │
                    ▼                                             │
┌───────┐    ┌───────────┐    ┌──────────────────┐    ┌──────────┴─┐
│ draft │───►│ designing │───►│ pending-approval │───►│  approved  │
└───────┘    └───────────┘    └──────────────────┘    └────────────┘
                                                            │
                                                            ▼
┌───────────┐    ┌───────────┐    ┌─────────┐    ┌───────────────┐
│ completed │◄───│  invoiced │◄───│delivered│◄───│in-production  │◄─┐
└───────────┘    └───────────┘    └─────────┘    └───────────────┘  │
                                       ▲                            │
                                       │         ┌─────────┐        │
                                       └─────────│ shipped │◄───────┤
                                                 └─────────┘        │
                                                                    │
                                                            ┌───────┴─┐
                                                            │ ordered │
                                                            └─────────┘

Any status ───► cancelled
```

**Status Descriptions:**

| Status | Description | Actions Available |
|--------|-------------|-------------------|
| draft | Project created, adding inlays | Add/remove inlays |
| designing | Proofs being created | Chat, create proofs |
| pending-approval | Proofs sent, awaiting approval | Approve/decline |
| approved | All inlays approved | Place order |
| ordered | Order placed, queued for production | - |
| in-production | Manufacturing in progress | Track milestones |
| shipped | All inlays shipped | Track delivery |
| delivered | Delivery confirmed | Create invoice |
| invoiced | Invoice sent | Pay |
| completed | Payment received | - |
| cancelled | Project cancelled | - |

### Proof Status Flow

```
┌─────────┐     ┌──────────┐
│ pending │────►│ approved │ (terminal)
└────┬────┘     └──────────┘
     │
     │          ┌──────────┐
     └─────────►│ declined │───► (new proof created)
                └──────────┘
                
     │          ┌────────────┐
     └─────────►│ superseded │ (when newer version exists)
                └────────────┘
```

**Rules:**
- A proof starts as `pending` when created
- `approved` is terminal - cannot be changed
- `declined` triggers feedback; designer creates new proof
- When a new proof is created, previous `pending` proofs become `superseded`

### Manufacturing Steps

```
ordered → materials-prep → cutting → fire-polish → packaging → shipped → delivered
```

**Key Characteristics:**
- Steps can move backward (via "revert" milestone events)
- Each transition creates an `inlay_milestone` record
- Progress is event-based, not a single status field
- Multiple blockers can exist per inlay

## Business Rules

### Ordering

#### Order Placement Requirements
```
✓ All inlays in the project must have an approved proof
✓ User must have "approver" or "admin" role
✓ Project must be in "approved" status
```

#### Price Locking
When an order is placed:
1. Create `order_snapshot` for each inlay
2. Snapshot includes: proof_id, price_group_id, price_cents, width, height
3. These values are immutable
4. Invoice uses snapshot prices, not current catalog prices

```go
// Pseudocode for order placement
func placeOrder(project *Project, user *DealershipUser) error {
    // Validate all inlays approved
    for _, inlay := range project.Inlays {
        if inlay.ApprovedProofID == nil {
            return fmt.Errorf("inlay %s not approved", inlay.Name)
        }
    }
    
    // Create snapshots (in transaction)
    for _, inlay := range project.Inlays {
        proof := getProof(inlay.ApprovedProofID)
        createOrderSnapshot(OrderSnapshot{
            ProjectID:    project.ID,
            InlayID:      inlay.ID,
            ProofID:      proof.ID,
            PriceGroupID: proof.PriceGroupID,
            PriceCents:   proof.PriceCents,
            Width:        proof.Width,
            Height:       proof.Height,
        })
    }
    
    // Update project
    project.Status = "ordered"
    project.OrderedAt = now()
    project.OrderedBy = user.ID
    
    // Initialize manufacturing
    for _, inlay := range project.Inlays {
        inlay.ManufacturingStep = "ordered"
        createMilestone(inlay.ID, "ordered", "entered")
    }
    
    // Notify
    createNotification("order_placed", project)
}
```

### Proofs

#### Price Group Assignment
- Price group is assigned at the **proof level**, not the inlay level
- A catalog item has a `default_price_group_id`
- Designer may change the price group based on:
  - Custom sizing (larger = more expensive)
  - Customization complexity
  - Special materials required

#### Proof Versioning
- Proofs are versioned per inlay: `(inlay_id, version_number)` is unique
- All versions are visible to the dealership
- Chat history carries across versions (single thread)

#### Proof-Chat Integration
- Proofs are "sent" within the chat
- When a proof is created:
  1. Create the `inlay_proof` record
  2. Create a chat message with `message_type = 'proof_sent'`
  3. Link: `proof.sent_in_chat_id = chat_message.id`

```go
func createAndSendProof(inlay *Inlay, proofData ProofInput, designer *InternalUser) error {
    // Get next version number
    versionNumber := getNextProofVersion(inlay.ID)
    
    // Create proof
    proof := InlayProof{
        InlayID:       inlay.ID,
        VersionNumber: versionNumber,
        PreviewURL:    proofData.PreviewURL,
        Width:         proofData.Width,
        Height:        proofData.Height,
        PriceGroupID:  proofData.PriceGroupID,
        Status:        "pending",
    }
    insertProof(&proof)
    
    // Mark previous pending proofs as superseded
    supersedePendingProofs(inlay.ID, proof.ID)
    
    // Create chat message
    chatMessage := InlayChat{
        InlayID:        inlay.ID,
        InternalUserID: designer.ID,
        MessageType:    "proof_sent",
        Message:        "New proof ready for review",
    }
    insertChat(&chatMessage)
    
    // Link proof to chat
    proof.SentInChatID = chatMessage.ID
    updateProof(&proof)
    
    // Update inlay preview
    inlay.PreviewURL = proof.PreviewURL
    updateInlay(inlay)
    
    // Notify dealership
    createNotification("proof_ready", inlay.ProjectID, inlay.ID)
    
    return nil
}
```

### Manufacturing

#### Milestone Events
Manufacturing progress is tracked via events, not a single status:

| Event Type | Meaning |
|------------|---------|
| entered | Inlay arrived at this step |
| exited | Inlay moved to next step |
| reverted | Inlay moved backward to this step |

**Example timeline:**
```
1. entered:ordered       (order placed)
2. exited:ordered        (starting materials)
3. entered:materials-prep
4. exited:materials-prep
5. entered:cutting
6. reverted:materials-prep  (problem found, going back)
7. exited:materials-prep
8. entered:cutting
... and so on
```

#### Current Step Derivation
The current manufacturing step is stored on `inlays.manufacturing_step` for query convenience, but the milestone history is the source of truth.

```go
func moveInlayToStep(inlay *Inlay, newStep string, user *InternalUser) error {
    // Check for hard blockers
    blockers := getActiveBlockers(inlay.ID)
    for _, b := range blockers {
        if b.BlockerType == "hard" && b.StepBlocked == inlay.ManufacturingStep {
            return fmt.Errorf("inlay has hard blocker: %s", b.Reason)
        }
    }
    
    // Determine event type
    eventType := "entered"
    if isBackwardMove(inlay.ManufacturingStep, newStep) {
        eventType = "reverted"
    }
    
    // Create milestone
    createMilestone(InlayMilestone{
        InlayID:     inlay.ID,
        Step:        newStep,
        EventType:   eventType,
        PerformedBy: user.ID,
    })
    
    // Update inlay
    inlay.ManufacturingStep = newStep
    updateInlay(inlay)
    
    // Notify if shipped or delivered
    if newStep == "shipped" || newStep == "delivered" {
        createNotification("inlay_step_changed", inlay.ProjectID, inlay.ID)
    }
    
    return nil
}
```

#### Blockers
Blockers indicate issues that need resolution:

| Type | Effect |
|------|--------|
| soft | Informational only, doesn't prevent progress |
| hard | Prevents moving to the next step |

**Multiple blockers:**
- An inlay can have multiple active blockers simultaneously
- Each blocker is resolved independently
- Hard blockers at the current step must all be resolved before moving

```go
func canMoveToNextStep(inlay *Inlay) (bool, string) {
    blockers := getActiveBlockers(inlay.ID)
    for _, b := range blockers {
        if b.BlockerType == "hard" && b.StepBlocked == inlay.ManufacturingStep {
            return false, fmt.Sprintf("Hard blocker: %s", b.Reason)
        }
    }
    return true, ""
}
```

### Users & Permissions

#### Multi-Tenancy
- Dealership users can only see their own dealership's data
- Every query must be scoped to `dealership_id`
- Internal users can see all dealerships' data

```go
// GOOD: Scoped query
func (m ProjectModel) GetAllForDealership(dealershipID int) ([]*Project, error) {
    query := postgres.SELECT(
        table.Projects.AllColumns,
    ).FROM(
        table.Projects,
    ).WHERE(
        table.Projects.DealershipID.EQ(postgres.Int(int64(dealershipID))),
    )
    // ...
}

// BAD: Unscoped query exposed to dealership users
func (m ProjectModel) GetAll() ([]*Project, error) {
    // This should only be used by internal users!
}
```

#### Dealership User Roles

| Role | Can Do |
|------|--------|
| viewer | View projects, chats, invoices |
| submitter | + Create projects, add inlays, chat |
| approver | + Approve/decline proofs, place orders |
| admin | + Manage users, pay invoices |

#### Internal User Roles

| Role | Can Do |
|------|--------|
| designer | Create proofs, respond to design chats |
| production | Manage kanban, create/resolve blockers |
| billing | Create invoices, mark paid |
| admin | Everything |

### Invoicing

#### Invoice Rules
- Invoices are 1:1 with projects
- Cannot create invoice until project is delivered
- Invoice line items are auto-populated from order snapshots
- Additional line items (shipping, fees) can be added manually
- Full payment only (no partial payments in MVP)

```go
func createInvoiceFromProject(project *Project) (*Invoice, error) {
    if project.Status != "delivered" {
        return nil, errors.New("project must be delivered before invoicing")
    }
    
    snapshots := getOrderSnapshots(project.ID)
    
    var subtotal int
    var lineItems []InvoiceLineItem
    
    for _, snap := range snapshots {
        inlay := getInlay(snap.InlayID)
        lineItems = append(lineItems, InvoiceLineItem{
            InlayID:        snap.InlayID,
            Description:    inlay.Name,
            Quantity:       1,
            UnitPriceCents: snap.PriceCents,
            TotalCents:     snap.PriceCents,
        })
        subtotal += snap.PriceCents
    }
    
    invoice := Invoice{
        ProjectID:     project.ID,
        InvoiceNumber: generateInvoiceNumber(),
        SubtotalCents: subtotal,
        TaxCents:      0, // Calculate if needed
        TotalCents:    subtotal,
        Status:        "draft",
    }
    
    // Insert in transaction with line items
    // ...
    
    return &invoice, nil
}
```

### Notifications

#### Event Types

| Event | Recipients | Description |
|-------|------------|-------------|
| proof_ready | Dealership users (approver+) | New proof available |
| proof_approved | Internal designers | Proof was approved |
| proof_declined | Internal designers | Proof was declined |
| order_placed | Internal production | New order in queue |
| inlay_step_changed | Dealership users | Inlay moved in manufacturing |
| inlay_blocked | Dealership users | Issue with inlay |
| inlay_unblocked | Dealership users | Issue resolved |
| project_shipped | Dealership users | Project shipped |
| project_delivered | Dealership users, internal billing | Ready for invoice |
| invoice_sent | Dealership users (admin) | Invoice available |
| payment_received | Dealership users | Payment confirmed |
| chat_message | Other party in chat | New message |

#### Notification Preferences
- Users can disable specific notification types
- Disabled notifications still appear in-app, just no email sent

```go
func sendNotification(eventType string, userID int, userType string, data NotificationData) error {
    // Create in-app notification (always)
    notification := Notification{
        EventType: eventType,
        Title:     data.Title,
        Body:      data.Body,
        ProjectID: data.ProjectID,
        InlayID:   data.InlayID,
    }
    
    if userType == "dealership" {
        notification.DealershipUserID = userID
    } else {
        notification.InternalUserID = userID
    }
    
    insertNotification(&notification)
    
    // Check email preference
    pref := getNotificationPreference(userID, userType, eventType)
    if pref == nil || pref.EmailEnabled {
        user := getUser(userID, userType)
        sendEmail(user.Email, data.Title, data.Body)
        notification.EmailSentAt = now()
        updateNotification(&notification)
    }
    
    return nil
}
```

## Catalog

### Catalog Items
- Have a unique `catalog_code` (e.g., "A-BRD-0003L")
- Default dimensions and minimum dimensions
- Default price group (can be overridden at proof level)
- Tags for searchability
- Multiple images (one primary)

### Catalog vs Custom Inlays

| Aspect | Catalog Inlay | Custom Inlay |
|--------|---------------|--------------|
| Reference | `catalog_item_id` | description + reference images |
| Initial dimensions | From catalog defaults | Customer's requested dimensions |
| Customization | `customization_notes` | Full custom design |
| Pricing basis | Catalog default + adjustments | Designer assessment |

## Future Considerations

### Graphical Editor (Post-MVP)
The data model supports future graphical editing:

```typescript
// inlay_proofs has these fields:
{
  scale_factor: 1.0,     // How much the design is scaled
  color_overrides: {}    // {"piece_id": "#hexcolor"}
}
```

When the graphical editor is built:
1. Start with catalog item's design asset
2. Apply `scale_factor` to resize
3. Apply `color_overrides` to recolor pieces
4. Generate `preview_url` from the result

### Per-Inlay Pricing (Post-MVP)
The `price_cents` field on `inlay_proofs` enables future per-inlay pricing:

- Currently: price derived from `price_group_id`
- Future: designer can set exact price via `price_cents`
- Order snapshot captures whichever is set
