package main

import (
	"fmt"
	"strings"
)

func generateDAO(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`package dao

import (
	"context"
	"%s/internal/models"
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx))
	})
}

func (d *DAO) DB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return d.db.WithContext(ctx)
}

`, projectName))

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		varName := toLowerCamelCase(table.Name)

		sb.WriteString(fmt.Sprintf(`
type %sDAO struct {
	*DAO
}

func (d *DAO) New%sDAO() *%sDAO {
	return &%sDAO{DAO: d}
}

func (dao *%sDAO) Create(ctx context.Context, %s *models.%s) error {
	return dao.DB(ctx).Create(%s).Error
}

func (dao *%sDAO) GetByID(ctx context.Context, id uint) (*models.%s, error) {
	var %s models.%s
	err := dao.DB(ctx).First(&%s, id).Error
	if err != nil {
		return nil, err
	}
	return &%s, nil
}

func (dao *%sDAO) List(ctx context.Context, page, pageSize int) ([]*models.%s, int64, error) {
	var %ss []*models.%s
	var total int64
	
	db := dao.DB(ctx).Model(&models.%s{})
	db.Count(&total)
	
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Find(&%ss).Error
	if err != nil {
		return nil, 0, err
	}
	return %ss, total, nil
}

func (dao *%sDAO) Update(ctx context.Context, %s *models.%s) error {
	return dao.DB(ctx).Save(%s).Error
}

func (dao *%sDAO) Delete(ctx context.Context, id uint) error {
	return dao.DB(ctx).Delete(&models.%s{}, id).Error
}
`,
			modelName,
			modelName, modelName, modelName,
			modelName, varName, modelName, varName,
			modelName, modelName, varName, modelName, varName, varName,
			modelName, modelName, varName, modelName, modelName, varName,
			modelName, modelName,
			modelName, varName, modelName, varName,
			modelName, modelName,
		))
	}

	return sb.String()
}

func generateDAOGen(tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(`package dao

import (
	"gorm.io/gen"
	"gorm.io/gorm"
)

type GenQuerier interface {
	GetByID(id int64) (gen.T, error)
	
	FindAll() ([]gen.T, error)
}

func GenerateDAO(db *gorm.DB) error {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./internal/dao/query",
		Mode:    gen.WithDefaultQuery,
	})
	
	g.UseDB(db)
	
	g.ApplyBasic(
`)

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		sb.WriteString(fmt.Sprintf("\t\tg.GenerateModelAs(\"%s\", \"%s\"),\n", table.Name, modelName))
	}

	sb.WriteString(`	)
	
	g.Execute()
	return nil
}
`)
	return sb.String()
}
