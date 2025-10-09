# Data Privacy & Protection Policy

## Table of Contents

1. [Introduction](#introduction)
2. [Data Collection](#data-collection)
3. [Data Processing](#data-processing)
4. [Data Storage](#data-storage)
5. [Data Retention](#data-retention)
6. [User Rights](#user-rights)
7. [Data Security](#data-security)
8. [Third-Party Services](#third-party-services)
9. [International Data Transfers](#international-data-transfers)
10. [Compliance](#compliance)

## Introduction

House Helper is committed to protecting the privacy and security of user data. This document outlines our technical implementation of data protection principles in compliance with GDPR, CCPA, and other privacy regulations.

## Data Collection

### Personal Data Collected

| Data Type | Purpose | Legal Basis | Retention |
|-----------|---------|-------------|-----------|
| Email address | Account creation, authentication | Contract | Account lifetime |
| Name | User identification, personalization | Contract | Account lifetime |
| Phone number (optional) | Two-factor authentication | Legitimate interest | Account lifetime |
| Profile picture (optional) | User identification | Consent | Account lifetime |
| Family membership | Access control, collaboration | Contract | Account lifetime |
| Task/chore data | Core functionality | Contract | Account lifetime |
| Points/rewards data | Gamification features | Contract | Account lifetime |
| Device tokens | Push notifications | Consent | Until revoked |
| Usage analytics | Service improvement | Legitimate interest | 90 days |
| Error logs | Debugging, service improvement | Legitimate interest | 30 days |

### Data Minimization

We implement data minimization by:
- Only collecting data necessary for functionality
- Making optional fields clearly marked
- Providing alternatives (e.g., username instead of real name)
- Regularly reviewing data collection practices

### Consent Management

```go
// User consent tracking
type UserConsent struct {
    UserID           string    `json:"user_id"`
    MarketingEmails  bool      `json:"marketing_emails"`
    Analytics        bool      `json:"analytics"`
    PushNotifications bool     `json:"push_notifications"`
    DataSharing      bool      `json:"data_sharing"`
    ConsentDate      time.Time `json:"consent_date"`
    ConsentVersion   string    `json:"consent_version"`
}

// Record consent
func (s *ConsentService) RecordConsent(userID string, consent UserConsent) error {
    consent.UserID = userID
    consent.ConsentDate = time.Now()
    consent.ConsentVersion = "1.0" // Track policy version
    
    return s.db.Create(&consent).Error
}
```

## Data Processing

### Purpose Limitation

Data is processed only for specified, explicit, and legitimate purposes:

1. **Account Management**: Authentication, authorization, profile management
2. **Core Functionality**: Task management, chore tracking, family collaboration
3. **Notifications**: Push notifications, email notifications
4. **Analytics**: Usage patterns, feature adoption, error tracking
5. **Support**: Customer support, bug investigation

### Processing Activities Record

```yaml
# Data processing activities (Article 30 GDPR)
processing_activities:
  - name: User Authentication
    purpose: Verify user identity and grant access
    legal_basis: Contract
    data_categories:
      - Email address
      - Password hash
      - Session tokens
    recipients: Internal systems only
    retention: Account lifetime
    security_measures:
      - Bcrypt password hashing
      - TLS encryption
      - Session token rotation
    
  - name: Push Notifications
    purpose: Send task reminders and updates
    legal_basis: Consent
    data_categories:
      - Device tokens
      - Notification preferences
    recipients:
      - Firebase Cloud Messaging
      - Apple Push Notification Service
    retention: Until consent withdrawn
    security_measures:
      - Encrypted transmission
      - Token expiration
      - Rate limiting
    
  - name: Analytics
    purpose: Improve service quality
    legal_basis: Legitimate interest
    data_categories:
      - Pseudonymized user ID
      - Feature usage
      - Error logs
    recipients: Internal analytics system
    retention: 90 days
    security_measures:
      - Pseudonymization
      - Aggregation
      - Access controls
```

### Automated Decision Making

House Helper does not use fully automated decision-making with legal or similarly significant effects. The points calculation system is:
- Transparent (rules visible to users)
- Configurable by family admins
- Subject to human review
- Not used for critical decisions

## Data Storage

### Storage Locations

| Data Type | Storage Location | Encryption |
|-----------|------------------|------------|
| User data | AWS RDS (us-east-1) | AES-256 at rest, TLS in transit |
| Session data | AWS ElastiCache Redis | In transit only |
| File uploads | AWS S3 (us-east-1) | AES-256 at rest |
| Event logs | AWS CloudWatch | AES-256 at rest |
| Backups | AWS S3 (us-east-1, us-west-2) | AES-256 at rest |

### Database Schema with Privacy Considerations

```sql
-- Users table with privacy controls
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(50),
    profile_picture_url TEXT,
    
    -- Privacy controls
    data_processing_consent BOOLEAN DEFAULT FALSE,
    marketing_consent BOOLEAN DEFAULT FALSE,
    analytics_consent BOOLEAN DEFAULT TRUE,
    
    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP, -- Soft delete for GDPR compliance
    
    -- Data subject access request tracking
    dsar_requested_at TIMESTAMP,
    dsar_completed_at TIMESTAMP
);

-- Audit log for data access
CREATE TABLE data_access_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    accessed_by UUID NOT NULL REFERENCES users(id),
    access_type VARCHAR(50) NOT NULL, -- read, update, delete, export
    resource_type VARCHAR(50) NOT NULL, -- user, task, family, etc.
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    
    INDEX idx_user_access (user_id, timestamp),
    INDEX idx_accessed_by (accessed_by, timestamp)
);

-- Consent history
CREATE TABLE consent_history (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    consent_type VARCHAR(50) NOT NULL,
    consent_given BOOLEAN NOT NULL,
    consent_version VARCHAR(20) NOT NULL,
    consented_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    INDEX idx_user_consent (user_id, consented_at)
);
```

### Data Pseudonymization

```go
// Pseudonymize user data for analytics
func (s *AnalyticsService) PseudonymizeUserData(userID string) string {
    // Use HMAC with secret key for consistent pseudonymization
    h := hmac.New(sha256.New, []byte(s.hmacSecret))
    h.Write([]byte(userID))
    return hex.EncodeToString(h.Sum(nil))
}

// Example usage
pseudoID := s.PseudonymizeUserData(user.ID)
analytics.TrackEvent(pseudoID, "task_completed", map[string]interface{}{
    "task_type": "chore",
    "duration": 15,
    // No personally identifiable information
})
```

## Data Retention

### Retention Periods

| Data Type | Retention Period | Deletion Method |
|-----------|------------------|-----------------|
| Active user data | Account lifetime | See user deletion process |
| Inactive accounts | 2 years after last login | Automated deletion |
| Task/chore data | Account lifetime | Cascading delete with user |
| Session data | 30 days | Automatic expiration |
| Audit logs | 7 years | Automated deletion |
| Analytics data | 90 days | Automated deletion |
| Error logs | 30 days | Automated deletion |
| Backups | 30 days | Automated deletion |

### Automated Deletion

```go
// Scheduled job for data retention compliance
func (s *RetentionService) CleanupExpiredData() error {
    ctx := context.Background()
    
    // Delete inactive accounts (no login for 2 years)
    inactiveThreshold := time.Now().AddDate(-2, 0, 0)
    result := s.db.Where("last_login_at < ? AND deleted_at IS NULL", inactiveThreshold).
        Update("deleted_at", time.Now())
    log.Printf("Soft deleted %d inactive accounts", result.RowsAffected)
    
    // Delete old analytics data (90 days)
    analyticsThreshold := time.Now().AddDate(0, 0, -90)
    result = s.db.Where("created_at < ?", analyticsThreshold).
        Delete(&AnalyticsEvent{})
    log.Printf("Deleted %d old analytics events", result.RowsAffected)
    
    // Delete old error logs (30 days)
    errorThreshold := time.Now().AddDate(0, 0, -30)
    result = s.db.Where("created_at < ?", errorThreshold).
        Delete(&ErrorLog{})
    log.Printf("Deleted %d old error logs", result.RowsAffected)
    
    // Permanently delete soft-deleted users (30 days after deletion)
    permanentDeleteThreshold := time.Now().AddDate(0, 0, -30)
    result = s.db.Unscoped().Where("deleted_at < ?", permanentDeleteThreshold).
        Delete(&User{})
    log.Printf("Permanently deleted %d users", result.RowsAffected)
    
    return nil
}
```

## User Rights

### Right to Access (Article 15 GDPR)

Users can request a copy of their personal data:

```go
// Data Subject Access Request (DSAR)
func (s *DSARService) ExportUserData(userID string) (*UserDataExport, error) {
    export := &UserDataExport{
        ExportDate: time.Now(),
        Format:     "JSON",
    }
    
    // User profile
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, err
    }
    export.User = user
    
    // Family memberships
    families, err := s.familyRepo.GetByUserID(userID)
    if err != nil {
        return nil, err
    }
    export.Families = families
    
    // Tasks and chores
    tasks, err := s.taskRepo.GetByUserID(userID)
    if err != nil {
        return nil, err
    }
    export.Tasks = tasks
    
    // Points history
    points, err := s.pointRepo.GetByUserID(userID)
    if err != nil {
        return nil, err
    }
    export.Points = points
    
    // Consent history
    consents, err := s.consentRepo.GetByUserID(userID)
    if err != nil {
        return nil, err
    }
    export.Consents = consents
    
    // Data access log
    accessLog, err := s.auditRepo.GetAccessLogByUserID(userID)
    if err != nil {
        return nil, err
    }
    export.AccessLog = accessLog
    
    // Log the DSAR
    s.auditRepo.LogDataAccess(userID, userID, "export", "user", userID)
    
    return export, nil
}
```

### Right to Rectification (Article 16 GDPR)

Users can update their personal data through the app or API:

```go
// Update user profile
func (s *UserService) UpdateProfile(userID string, updates UserProfileUpdate) error {
    // Validate input
    if err := s.validator.Validate(updates); err != nil {
        return err
    }
    
    // Update user
    err := s.userRepo.Update(userID, updates)
    if err != nil {
        return err
    }
    
    // Log the update
    s.auditRepo.LogDataAccess(userID, userID, "update", "user", userID)
    
    return nil
}
```

### Right to Erasure (Article 17 GDPR)

Users can request deletion of their account and data:

```go
// Delete user account and associated data
func (s *UserService) DeleteUser(userID string, reason string) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // 1. Anonymize personal data
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    anonymizedData := map[string]interface{}{
        "email":              fmt.Sprintf("deleted-%s@deleted.local", userID),
        "name":               "Deleted User",
        "phone":              nil,
        "profile_picture_url": nil,
        "password_hash":      "DELETED",
    }
    
    if err := tx.Model(&User{}).Where("id = ?", userID).Updates(anonymizedData).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 2. Delete device tokens (stops notifications)
    if err := tx.Where("user_id = ?", userID).Delete(&DeviceToken{}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 3. Soft delete user record
    if err := tx.Model(&User{}).Where("id = ?", userID).Update("deleted_at", time.Now()).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 4. Remove from families (transfer ownership or delete if sole member)
    families, err := s.familyRepo.GetByUserID(userID)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    for _, family := range families {
        if family.OwnerID == userID {
            // Transfer ownership or delete family
            members, _ := s.familyRepo.GetMembers(family.ID)
            if len(members) > 1 {
                // Transfer to another member
                newOwner := members[0]
                if err := tx.Model(&Family{}).Where("id = ?", family.ID).Update("owner_id", newOwner.ID).Error; err != nil {
                    tx.Rollback()
                    return err
                }
            } else {
                // Delete family if sole member
                if err := tx.Where("id = ?", family.ID).Delete(&Family{}).Error; err != nil {
                    tx.Rollback()
                    return err
                }
            }
        }
        
        // Remove membership
        if err := tx.Where("family_id = ? AND user_id = ?", family.ID, userID).Delete(&FamilyMember{}).Error; err != nil {
            tx.Rollback()
            return err
        }
    }
    
    // 5. Anonymize task assignments
    if err := tx.Model(&Task{}).Where("assigned_to = ?", userID).Update("assigned_to", nil).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 6. Keep audit trail but anonymize
    if err := tx.Model(&DataAccessLog{}).Where("user_id = ?", userID).
        Updates(map[string]interface{}{"user_id": nil}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 7. Log the deletion
    if err := s.auditRepo.LogDeletion(userID, reason); err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    
    // 8. Schedule permanent deletion after 30 days
    s.schedulePermandentDeletion(userID, time.Now().AddDate(0, 0, 30))
    
    return nil
}
```

### Right to Data Portability (Article 20 GDPR)

Users can download their data in a machine-readable format:

```go
// Export user data in JSON format
func (s *ExportService) ExportToJSON(userID string) ([]byte, error) {
    data, err := s.dsarService.ExportUserData(userID)
    if err != nil {
        return nil, err
    }
    
    return json.MarshalIndent(data, "", "  ")
}

// Export user data in CSV format
func (s *ExportService) ExportToCSV(userID string) ([]byte, error) {
    data, err := s.dsarService.ExportUserData(userID)
    if err != nil {
        return nil, err
    }
    
    // Convert to CSV
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    
    // Write tasks
    writer.Write([]string{"Type", "Title", "Description", "Status", "Due Date", "Points"})
    for _, task := range data.Tasks {
        writer.Write([]string{
            "Task",
            task.Title,
            task.Description,
            task.Status,
            task.DueDate.Format("2006-01-02"),
            strconv.Itoa(task.Points),
        })
    }
    
    writer.Flush()
    return buf.Bytes(), nil
}
```

### Right to Object (Article 21 GDPR)

Users can object to certain data processing:

```go
// Object to marketing communications
func (s *ConsentService) ObjectToMarketing(userID string) error {
    return s.UpdateConsent(userID, ConsentUpdate{
        MarketingEmails:   false,
        AnalyticsConsent:  nil, // Don't change
    })
}

// Object to analytics
func (s *ConsentService) ObjectToAnalytics(userID string) error {
    return s.UpdateConsent(userID, ConsentUpdate{
        MarketingEmails:  nil, // Don't change
        AnalyticsConsent: false,
    })
}
```

### Right to Restriction of Processing (Article 18 GDPR)

Users can request restriction of processing in certain circumstances:

```go
// Restrict user data processing
func (s *UserService) RestrictProcessing(userID string, reason string) error {
    return s.userRepo.Update(userID, map[string]interface{}{
        "processing_restricted": true,
        "restriction_reason":    reason,
        "restricted_at":         time.Now(),
    })
}

// Check if processing is restricted
func (s *UserService) CanProcess(userID string) (bool, error) {
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return false, err
    }
    
    return !user.ProcessingRestricted, nil
}
```

## Data Security

See [SECURITY.md](./SECURITY.md) for comprehensive security measures.

### Key Security Measures

1. **Encryption**: AES-256 at rest, TLS 1.3 in transit
2. **Access Control**: Role-based access control (RBAC)
3. **Authentication**: Bcrypt password hashing, JWT tokens
4. **Authorization**: Principle of least privilege
5. **Monitoring**: Audit logs, security alerts
6. **Backups**: Encrypted, tested regularly
7. **Incident Response**: Documented procedures

## Third-Party Services

| Service | Purpose | Data Shared | Privacy Policy |
|---------|---------|-------------|----------------|
| AWS | Infrastructure hosting | All application data | [AWS Privacy](https://aws.amazon.com/privacy/) |
| Firebase Cloud Messaging | Push notifications | Device tokens, notifications | [Firebase Privacy](https://firebase.google.com/support/privacy) |
| Apple Push Notification | iOS push notifications | Device tokens, notifications | [Apple Privacy](https://www.apple.com/privacy/) |
| Temporal.io | Workflow orchestration | Task scheduling data | [Temporal Privacy](https://temporal.io/privacy-policy) |
| Sentry (optional) | Error tracking | Error logs, stack traces | [Sentry Privacy](https://sentry.io/privacy/) |

### Data Processing Agreements (DPAs)

We have Data Processing Agreements in place with all third-party processors that handle personal data, ensuring:
- GDPR compliance
- Appropriate security measures
- Sub-processor management
- Data breach notification
- Audit rights

## International Data Transfers

### Data Transfer Mechanisms

- **AWS Regions**: Primary region us-east-1, backup us-west-2 (both in USA)
- **Legal Basis**: Standard Contractual Clauses (SCCs)
- **EU Users**: Data stored in EU regions if required
- **Adequacy Decisions**: Follow EC adequacy decisions

### Cross-Border Transfer Safeguards

```go
// Configure data residency based on user location
func (s *StorageService) GetStorageRegion(user *User) string {
    switch user.Country {
    case "DE", "FR", "IT", "ES", "NL", "BE", "AT", "SE", "DK", "FI":
        // EU users → EU region
        return "eu-central-1"
    case "GB":
        // UK users → UK region
        return "eu-west-2"
    default:
        // Other users → US region
        return "us-east-1"
    }
}
```

## Compliance

### GDPR Compliance

- [x] Lawful basis for processing
- [x] Data minimization
- [x] Purpose limitation
- [x] Accuracy
- [x] Storage limitation
- [x] Integrity and confidentiality
- [x] Accountability
- [x] Privacy by design and default
- [x] Data Protection Impact Assessment (DPIA) completed
- [x] Data Protection Officer (DPO) appointed

### CCPA Compliance

- [x] Right to know
- [x] Right to delete
- [x] Right to opt-out
- [x] Right to non-discrimination
- [x] Privacy notice
- [x] Do Not Sell My Personal Information

### Other Regulations

- COPPA: No users under 13 without parental consent
- FERPA: Educational data protected if applicable
- HIPAA: Not applicable (no health data)

### Audits and Certifications

- Annual security audit
- SOC 2 Type II (planned)
- ISO 27001 certification (planned)
- GDPR compliance audit

## Contact

- **Privacy Questions**: privacy@house-helper.com
- **Data Protection Officer**: dpo@house-helper.com
- **Data Subject Requests**: dsar@house-helper.com

## Updates

This privacy policy is reviewed and updated:
- Annually
- When regulations change
- When processing activities change
- When technical measures change

Users are notified of material changes via email and in-app notification.

---

**Last Updated**: 2024-01-15  
**Version**: 1.0  
**Next Review**: 2025-01-15

## License

Copyright © 2024 House Helper. All rights reserved.
