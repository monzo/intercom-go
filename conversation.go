package intercom

// ConversationService handles interactions with the API through an ConversationRepository.
type ConversationService struct {
	Repository ConversationRepository
}

// ConversationList is a list of Conversations
type ConversationList struct {
	Pages         PageParams     `json:"pages"`
	Conversations []Conversation `json:"conversations"`
}

// A Conversation represents a conversation between users and admins in Intercom.
type Conversation struct {
	ID                  string               `json:"id"`
	CreatedAt           int64                `json:"created_at"`
	UpdatedAt           int64                `json:"updated_at"`
	User                User                 `json:"user"`
	Assignee            Admin                `json:"assignee"`
	Open                bool                 `json:"open"`
	Read                bool                 `json:"read"`
	ConversationMessage ConversationMessage  `json:"conversation_message"`
	ConversationParts   ConversationPartList `json:"conversation_parts"`
	TagList             *TagList             `json:"tags"`
	ConversationRating  ConversationRating   `json:"conversation_rating"`
}

type Customer struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Teammate struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// A ConversationRating is the customer satisfaction rating result
type ConversationRating struct {
	Rating    int64    `json:"rating"`
	Remark    string   `json:"remark"`
	CreatedAt int64    `json:"created_at"`
	Customer  Customer `json:"customer"`
	Teammate  Teammate `json:"teammate"`
}

// A ConversationMessage is the message that started the conversation rendered for presentation
type ConversationMessage struct {
	Subject     string         `json:"subject"`
	Body        string         `json:"body"`
	Author      MessageAddress `json:"author"`
	URL         string         `json:"url"`
	Attachments []Attachment   `json:"attachments"`
}

// A ConversationPartList lists the subsequent Conversation Parts
type ConversationPartList struct {
	Parts []ConversationPart `json:"conversation_parts"`
}

// A ConversationPart is a Reply, Note, or Assignment to a Conversation
type ConversationPart struct {
	ID          string         `json:"id"`
	PartType    string         `json:"part_type"`
	Body        string         `json:"body"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	NotifiedAt  int64          `json:"notified_at"`
	AssignedTo  Admin          `json:"assigned_to"`
	Author      MessageAddress `json:"author"`
	Attachments []Attachment   `json:"attachments"`
}

type Attachment struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// The state of Conversations to query
// SHOW_ALL shows all conversations,
// SHOW_OPEN shows only open conversations (only valid for Admin Conversation queries)
// SHOW_CLOSED shows only closed conversations (only valid for Admin Conversation queries)
// SHOW_UNREAD shows only unread conversations (only valid for User Conversation queries)
type ConversationListState int

const (
	SHOW_ALL ConversationListState = iota
	SHOW_OPEN
	SHOW_CLOSED
	SHOW_UNREAD
)

// List all Conversations
func (c *ConversationService) ListAll(pageParams PageParams) (ConversationList, error) {
	return c.Repository.list(conversationListParams{PageParams: pageParams})
}

// List Conversations by Admin
func (c *ConversationService) ListByAdmin(adminID string, orderBy string, sort string, state ConversationListState, pageParams PageParams) (ConversationList, error) {
	params := conversationListParams{
		PageParams: pageParams,
		Type:       "admin",
		AdminID:    adminID,
		Order:      orderBy,
		Sort:       sort,
	}
	if state == SHOW_OPEN {
		params.Open = Bool(true)
		params.State = "open"
	}
	if state == SHOW_CLOSED {
		params.Open = Bool(false)
		params.State = "closed"
	}
	return c.Repository.list(params)
}

// List Conversations by User
func (c *ConversationService) ListByUser(user *User, state ConversationListState, pageParams PageParams) (ConversationList, error) {
	params := conversationListParams{
		PageParams:     pageParams,
		Type:           "user",
		IntercomUserID: user.ID,
		UserID:         user.UserID,
		Email:          user.Email,
	}
	if state == SHOW_UNREAD {
		params.Unread = Bool(true)
	}
	return c.Repository.list(params)
}

// Find Conversation by conversation id
func (c *ConversationService) Find(id string) (Conversation, error) {
	return c.Repository.find(id)
}

// Mark Conversation as read (by a User)
func (c *ConversationService) MarkRead(id string) (Conversation, error) {
	return c.Repository.read(id)
}

func (c *ConversationService) Reply(id string, author MessagePerson, replyType ReplyType, body string) (Conversation, error) {
	return c.reply(id, author, replyType, body, nil)
}

// Reply to a Conversation by id
func (c *ConversationService) ReplyWithAttachmentURLs(id string, author MessagePerson, replyType ReplyType, body string, attachmentURLs []string) (Conversation, error) {
	return c.reply(id, author, replyType, body, attachmentURLs)
}

func (c *ConversationService) reply(id string, author MessagePerson, replyType ReplyType, body string, attachmentURLs []string) (Conversation, error) {
	addr := author.MessageAddress()
	reply := Reply{
		Type:           addr.Type,
		ReplyType:      replyType.String(),
		Body:           body,
		AttachmentURLs: attachmentURLs,
	}
	if addr.Type == "admin" {
		reply.AdminID = addr.ID
	} else {
		reply.IntercomID = addr.ID
		reply.UserID = addr.UserID
		reply.Email = addr.Email
	}
	return c.Repository.reply(id, &reply)
}

// Assign a Conversation to an Admin
func (c *ConversationService) Assign(id string, assigner, assignee *Admin) (Conversation, error) {
	assignerAddr := assigner.MessageAddress()
	assigneeAddr := assignee.MessageAddress()
	reply := Reply{
		Type:       "admin",
		ReplyType:  CONVERSATION_ASSIGN.String(),
		AdminID:    assignerAddr.ID,
		AssigneeID: assigneeAddr.ID,
	}
	return c.Repository.reply(id, &reply)
}

// Open a Conversation (without a body)
func (c *ConversationService) Open(id string, opener *Admin) (Conversation, error) {
	return c.reply(id, opener, CONVERSATION_OPEN, "", nil)
}

// Close a Conversation (without a body)
func (c *ConversationService) Close(id string, closer *Admin) (Conversation, error) {
	return c.reply(id, closer, CONVERSATION_CLOSE, "", nil)
}

type conversationListParams struct {
	PageParams
	Type           string `url:"type,omitempty"`
	AdminID        string `url:"admin_id,omitempty"`
	IntercomUserID string `url:"intercom_user_id,omitempty"`
	UserID         string `url:"user_id,omitempty"`
	Email          string `url:"email,omitempty"`
	Open           *bool  `url:"open,omitempty"`
	Unread         *bool  `url:"unread,omitempty"`
	DisplayAs      string `url:"display_as,omitempty"`
	Order          string `url:"order,omitempty"`
	Sort           string `url:"sort,omitempty"`
	State          string `url:"state,omitempty"`
}
