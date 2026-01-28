# Te Reo Bot - Product Constitution

## Core Mission

**WHY**: Promote and preserve te reo M\u0101ori language through daily exposure on social media.

**WHAT**: Automated daily posts of M\u0101ori words with English meanings and culturally relevant images.

**WHO**:
- M\u0101ori language learners globally
- Anyone interested in M\u0101ori culture and language
- Social media users on Mastodon and Twitter

## Key Features

### Current Capabilities

**Daily Word Publishing**
- 366 unique M\u0101ori words (one per day, including leap years)
- English translations and meanings
- Cultural context where relevant
- High-quality images with attribution

**Multi-Platform Distribution**
- Mastodon integration
- Twitter integration
- Scheduled automated posting
- API-driven architecture for future platform additions

**Content Management**
- SQLite database for word storage
- Dictionary validation (366 unique entries)
- JSON export for backward compatibility
- Image storage via Google Cloud Storage

**Word Selection**
- Day-of-year based selection (deterministic)
- Leap year support (day 366)
- Consistent timing across years

## Target Audience

### Primary Users
- **Language Learners**: Daily exposure to M\u0101ori vocabulary
- **Educators**: Classroom resource for teaching te reo M\u0101ori
- **Cultural Enthusiasts**: Regular engagement with M\u0101ori language

### Secondary Users
- **Developers**: Example of scheduled social media automation
- **Content Curators**: Template for educational bot implementations

## Use Cases

1. **Daily Learning**: Users follow the account for regular language exposure
2. **Educational Resource**: Teachers share daily posts in classrooms
3. **Cultural Engagement**: Maintaining connection to M\u0101ori language and culture
4. **Social Media Presence**: Automated, consistent content delivery

## Important Scope Decisions

### In Scope

**Current (v1.0)**
- 366 M\u0101ori words with English translations
- Daily automated posting
- Mastodon and Twitter platforms
- Image attachments with attribution
- Word database management via CLI
- Data validation and integrity checks

**Future Considerations**
- Additional social media platforms (Bluesky, Threads)
- Enhanced cultural context in posts
- Pronunciation guides (text/audio)
- Word categories or themes
- Analytics and engagement metrics

### Out of Scope (PERMANENTLY)

**Multi-Language Support**
- Only M\u0101ori words with English meanings
- No internationalization planned
- English is the interface language

**User-Generated Content**
- Curated content only
- No public word submissions
- Maintained by repository owners

**Real-Time Interactive Features**
- No chatbot functionality
- No interactive quizzes or games
- Scheduled posts only, not conversational

**Mobile Applications**
- Social media platform access only
- No native iOS/Android apps
- No dedicated mobile UI

**Monetization**
- Free public service
- No advertisements
- No premium features or subscriptions

### Out of Scope (v1.0)

**Advanced Analytics**
- Basic posting metrics only
- No detailed engagement analysis
- No A/B testing or optimization

**Content Variations**
- Single format per day
- No multiple post variations
- No personalization

## Success Criteria

### User Adoption
- Growing follower count on both platforms
- Regular engagement (likes, shares, comments)
- Positive community feedback

### Content Quality
- All 366 words published accurately
- Culturally appropriate images
- Proper attributions maintained
- No posting failures or errors

### Reliability
- Daily posts without manual intervention
- 99.9% uptime for posting service
- Automated recovery from failures

### Cultural Impact
- Respectful representation of M\u0101ori language
- Positive feedback from M\u0101ori community
- Educational value recognized by users

## Roadmap Phases

### Phase 1: Foundation (Complete)
- HTTP server for scheduled posts
- Twitter and Mastodon integration
- JSON-based dictionary management
- Basic error handling and logging

### Phase 2: Database Migration (Current)
- SQLite storage for word management
- CLI tool for dictionary operations
- Data validation and integrity
- Backward-compatible JSON generation

### Phase 3: Enhanced Content (Future)
- Richer cultural context
- Pronunciation guides
- Word categories/themes
- Improved image selection

### Phase 4: Platform Expansion (Future)
- Additional social media platforms
- Analytics and insights
- Community engagement features
- Documentation and guides for contributors
