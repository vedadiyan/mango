package generators

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// +--------------------------------------------------+
// | Begin Type Definition For: Export
// +--------------------------------------------------+

type Export bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: Address
// +--------------------------------------------------+

var _ExportAddressName = db.FieldName[ExportAddressSpecs]()

func ExportAddressName() string {
	return _ExportAddressName
}

type ExportAddressSpecs[T any] struct {
	Address *T `bson:"address,omitempty"`
}

func (x Export) ExportAddress(v ExportAddress) Export {
	return append(x, bson.E{Key: ExportAddressName(), Value: v})
}

func ExportAddressBsonE(v ExportAddress) bson.E {
	return bson.E{Key: ExportAddressName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Address
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Details
// +--------------------------------------------------+

var _ExportDetailsName = db.FieldName[ExportDetailsSpecs]()

func ExportDetailsName() string {
	return _ExportDetailsName
}

type ExportDetailsSpecs[T any] struct {
	Details T `bson:"details"`
}

func (x Export) ExportDetails(v ExportDetails) Export {
	return append(x, bson.E{Key: ExportDetailsName(), Value: v})
}

func ExportDetailsBsonE(v ExportDetails) bson.E {
	return bson.E{Key: ExportDetailsName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Details
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: FirstName
// +--------------------------------------------------+

var _ExportFirstNameName = db.FieldName[ExportFirstNameSpecs]()

func ExportFirstNameName() string {
	return _ExportFirstNameName
}

type ExportFirstNameSpecs struct {
	FirstName *string `bson:"first_name,omitempty"`
}

func (x Export) ExportFirstName(v *string) Export {
	return append(x, bson.E{Key: ExportFirstNameName(), Value: v})
}

func ExportFirstNameBsonE(v *string) bson.E {
	return bson.E{Key: ExportFirstNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: FirstName
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Hobbie
// +--------------------------------------------------+

var _ExportHobbieName = db.FieldName[ExportHobbieSpecs]()

func ExportHobbieName() string {
	return _ExportHobbieName
}

type ExportHobbieSpecs struct {
	Hobbie *string `bson:"hobbie,omitempty"`
}

func (x Export) ExportHobbie(v *string) Export {
	return append(x, bson.E{Key: ExportHobbieName(), Value: v})
}

func ExportHobbieBsonE(v *string) bson.E {
	return bson.E{Key: ExportHobbieName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Hobbie
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Hobbie2
// +--------------------------------------------------+

var _ExportHobbie2Name = db.FieldName[ExportHobbie2Specs]()

func ExportHobbie2Name() string {
	return _ExportHobbie2Name
}

type ExportHobbie2Specs struct {
	Hobbie2 *string `bson:"hobbie2,omitempty"`
}

func (x Export) ExportHobbie2(v *string) Export {
	return append(x, bson.E{Key: ExportHobbie2Name(), Value: v})
}

func ExportHobbie2BsonE(v *string) bson.E {
	return bson.E{Key: ExportHobbie2Name(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Hobbie2
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: LastName
// +--------------------------------------------------+

var _ExportLastNameName = db.FieldName[ExportLastNameSpecs]()

func ExportLastNameName() string {
	return _ExportLastNameName
}

type ExportLastNameSpecs struct {
	LastName *string `bson:"last_name,omitempty"`
}

func (x Export) ExportLastName(v *string) Export {
	return append(x, bson.E{Key: ExportLastNameName(), Value: v})
}

func ExportLastNameBsonE(v *string) bson.E {
	return bson.E{Key: ExportLastNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: LastName
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Products
// +--------------------------------------------------+

var _ExportProductsName = db.FieldName[ExportProductsSpecs]()

func ExportProductsName() string {
	return _ExportProductsName
}

type ExportProductsSpecs[T any] struct {
	Products []T `bson:"products"`
}

func (x Export) ExportProducts(v ExportProducts) Export {
	return append(x, bson.E{Key: ExportProductsName(), Value: v})
}

func ExportProductsBsonE(v ExportProducts) bson.E {
	return bson.E{Key: ExportProductsName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Products
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Responsibility
// +--------------------------------------------------+

var _ExportResponsibilityName = db.FieldName[ExportResponsibilitySpecs]()

func ExportResponsibilityName() string {
	return _ExportResponsibilityName
}

type ExportResponsibilitySpecs struct {
	Responsibility *time.Time `bson:"responsibility,omitempty,opaque"`
}

func (x Export) ExportResponsibility(v *time.Time) Export {
	return append(x, bson.E{Key: ExportResponsibilityName(), Value: v})
}

func ExportResponsibilityBsonE(v *time.Time) bson.E {
	return bson.E{Key: ExportResponsibilityName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Responsibility
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Role
// +--------------------------------------------------+

var _ExportRoleName = db.FieldName[ExportRoleSpecs]()

func ExportRoleName() string {
	return _ExportRoleName
}

type ExportRoleSpecs struct {
	Role *string `bson:"role,omitempty"`
}

func (x Export) ExportRole(v *string) Export {
	return append(x, bson.E{Key: ExportRoleName(), Value: v})
}

func ExportRoleBsonE(v *string) bson.E {
	return bson.E{Key: ExportRoleName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Role
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Skills
// +--------------------------------------------------+

var _ExportSkillsName = db.FieldName[ExportSkillsSpecs]()

func ExportSkillsName() string {
	return _ExportSkillsName
}

type ExportSkillsSpecs struct {
	Skills *string `bson:"skills,omitempty"`
}

func (x Export) ExportSkills(v *string) Export {
	return append(x, bson.E{Key: ExportSkillsName(), Value: v})
}

func ExportSkillsBsonE(v *string) bson.E {
	return bson.E{Key: ExportSkillsName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Skills
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Subscription
// +--------------------------------------------------+

var _ExportSubscriptionName = db.FieldName[ExportSubscriptionSpecs]()

func ExportSubscriptionName() string {
	return _ExportSubscriptionName
}

type ExportSubscriptionSpecs struct {
	Subscription *ExportSubscription `bson:"subscription,omitempty"`
}

// +--------------------------------------------------+
// | End Type Definition For: Export
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportAddress
// +--------------------------------------------------+

type ExportAddress bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: PostalCode
// +--------------------------------------------------+

var _ExportAddressPostalCodeName = db.FieldName[ExportAddressPostalCodeSpecs]()

func ExportAddressPostalCodeName() string {
	return _ExportAddressPostalCodeName
}

type ExportAddressPostalCodeSpecs struct {
	PostalCode *string `bson:"postal_code,omitempty"`
}

func (x ExportAddress) ExportAddressPostalCode(v *string) ExportAddress {
	return append(x, bson.E{Key: ExportAddressPostalCodeName(), Value: v})
}

func ExportAddressPostalCodeBsonE(v *string) bson.E {
	return bson.E{Key: ExportAddressPostalCodeName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: PostalCode
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Street
// +--------------------------------------------------+

var _ExportAddressStreetName = db.FieldName[ExportAddressStreetSpecs]()

func ExportAddressStreetName() string {
	return _ExportAddressStreetName
}

type ExportAddressStreetSpecs struct {
	Street *string `bson:"street,omitempty"`
}

func (x ExportAddress) ExportAddressStreet(v *string) ExportAddress {
	return append(x, bson.E{Key: ExportAddressStreetName(), Value: v})
}

func ExportAddressStreetBsonE(v *string) bson.E {
	return bson.E{Key: ExportAddressStreetName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Street
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportAddress
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportDetails
// +--------------------------------------------------+

type ExportDetails bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: CustomerName
// +--------------------------------------------------+

var _ExportDetailsCustomerNameName = db.FieldName[ExportDetailsCustomerNameSpecs]()

func ExportDetailsCustomerNameName() string {
	return _ExportDetailsCustomerNameName
}

type ExportDetailsCustomerNameSpecs struct {
	CustomerName *string `bson:"customer_name,omitempty"`
}

func (x ExportDetails) ExportDetailsCustomerName(v *string) ExportDetails {
	return append(x, bson.E{Key: ExportDetailsCustomerNameName(), Value: v})
}

func ExportDetailsCustomerNameBsonE(v *string) bson.E {
	return bson.E{Key: ExportDetailsCustomerNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: CustomerName
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Email
// +--------------------------------------------------+

var _ExportDetailsEmailName = db.FieldName[ExportDetailsEmailSpecs]()

func ExportDetailsEmailName() string {
	return _ExportDetailsEmailName
}

type ExportDetailsEmailSpecs struct {
	Email *string `bson:"email,omitempty"`
}

func (x ExportDetails) ExportDetailsEmail(v *string) ExportDetails {
	return append(x, bson.E{Key: ExportDetailsEmailName(), Value: v})
}

func ExportDetailsEmailBsonE(v *string) bson.E {
	return bson.E{Key: ExportDetailsEmailName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Email
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Legal
// +--------------------------------------------------+

var _ExportDetailsLegalName = db.FieldName[ExportDetailsLegalSpecs]()

func ExportDetailsLegalName() string {
	return _ExportDetailsLegalName
}

type ExportDetailsLegalSpecs[T any] struct {
	Legal *T `bson:"legal,omitempty"`
}

func (x ExportDetails) ExportDetailsLegal(v ExportDetailsLegal) ExportDetails {
	return append(x, bson.E{Key: ExportDetailsLegalName(), Value: v})
}

func ExportDetailsLegalBsonE(v ExportDetailsLegal) bson.E {
	return bson.E{Key: ExportDetailsLegalName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Legal
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportDetails
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportDetailsLegal
// +--------------------------------------------------+

type ExportDetailsLegal bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: CompanyName
// +--------------------------------------------------+

var _ExportDetailsLegalCompanyNameName = db.FieldName[ExportDetailsLegalCompanyNameSpecs]()

func ExportDetailsLegalCompanyNameName() string {
	return _ExportDetailsLegalCompanyNameName
}

type ExportDetailsLegalCompanyNameSpecs struct {
	CompanyName string `bson:"company_name"`
}

func (x ExportDetailsLegal) ExportDetailsLegalCompanyName(v string) ExportDetailsLegal {
	return append(x, bson.E{Key: ExportDetailsLegalCompanyNameName(), Value: v})
}

func ExportDetailsLegalCompanyNameBsonE(v string) bson.E {
	return bson.E{Key: ExportDetailsLegalCompanyNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: CompanyName
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: InnerField
// +--------------------------------------------------+

var _ExportDetailsLegalInnerFieldName = db.FieldName[ExportDetailsLegalInnerFieldSpecs]()

func ExportDetailsLegalInnerFieldName() string {
	return _ExportDetailsLegalInnerFieldName
}

type ExportDetailsLegalInnerFieldSpecs[T any] struct {
	InnerField *T `bson:"inner_field,omitempty"`
}

func (x ExportDetailsLegal) ExportDetailsLegalInnerField(v ExportDetailsLegalInnerField) ExportDetailsLegal {
	return append(x, bson.E{Key: ExportDetailsLegalInnerFieldName(), Value: v})
}

func ExportDetailsLegalInnerFieldBsonE(v ExportDetailsLegalInnerField) bson.E {
	return bson.E{Key: ExportDetailsLegalInnerFieldName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: InnerField
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: TaxId
// +--------------------------------------------------+

var _ExportDetailsLegalTaxIdName = db.FieldName[ExportDetailsLegalTaxIdSpecs]()

func ExportDetailsLegalTaxIdName() string {
	return _ExportDetailsLegalTaxIdName
}

type ExportDetailsLegalTaxIdSpecs struct {
	TaxId string `bson:"tax_id"`
}

func (x ExportDetailsLegal) ExportDetailsLegalTaxId(v string) ExportDetailsLegal {
	return append(x, bson.E{Key: ExportDetailsLegalTaxIdName(), Value: v})
}

func ExportDetailsLegalTaxIdBsonE(v string) bson.E {
	return bson.E{Key: ExportDetailsLegalTaxIdName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: TaxId
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportDetailsLegal
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportDetailsLegalInnerField
// +--------------------------------------------------+

type ExportDetailsLegalInnerField bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: Id
// +--------------------------------------------------+

var _ExportDetailsLegalInnerFieldIdName = db.FieldName[ExportDetailsLegalInnerFieldIdSpecs]()

func ExportDetailsLegalInnerFieldIdName() string {
	return _ExportDetailsLegalInnerFieldIdName
}

type ExportDetailsLegalInnerFieldIdSpecs struct {
	Id *int32 `bson:"id,omitempty"`
}

func (x ExportDetailsLegalInnerField) ExportDetailsLegalInnerFieldId(v *int32) ExportDetailsLegalInnerField {
	return append(x, bson.E{Key: ExportDetailsLegalInnerFieldIdName(), Value: v})
}

func ExportDetailsLegalInnerFieldIdBsonE(v *int32) bson.E {
	return bson.E{Key: ExportDetailsLegalInnerFieldIdName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Id
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportDetailsLegalInnerField
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportProducts
// +--------------------------------------------------+

type ExportProducts bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: Id
// +--------------------------------------------------+

var _ExportProductsIdName = db.FieldName[ExportProductsIdSpecs]()

func ExportProductsIdName() string {
	return _ExportProductsIdName
}

type ExportProductsIdSpecs struct {
	Id *int32 `bson:"id,omitempty"`
}

func (x ExportProducts) ExportProductsId(v *int32) ExportProducts {
	return append(x, bson.E{Key: ExportProductsIdName(), Value: v})
}

func ExportProductsIdBsonE(v *int32) bson.E {
	return bson.E{Key: ExportProductsIdName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Id
// +--------------------------------------------------+

// +--------------------------------------------------+
// |    Begin Field Definition For: Name
// +--------------------------------------------------+

var _ExportProductsNameName = db.FieldName[ExportProductsNameSpecs]()

func ExportProductsNameName() string {
	return _ExportProductsNameName
}

type ExportProductsNameSpecs struct {
	Name *string `bson:"name,omitempty"`
}

func (x ExportProducts) ExportProductsName(v *string) ExportProducts {
	return append(x, bson.E{Key: ExportProductsNameName(), Value: v})
}

func ExportProductsNameBsonE(v *string) bson.E {
	return bson.E{Key: ExportProductsNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Name
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportProducts
// +--------------------------------------------------+
