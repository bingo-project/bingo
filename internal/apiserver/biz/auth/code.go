// ABOUTME: Unified verification code business logic for multi-auth system.
// ABOUTME: Handles code generation, storage, sending (email/SMS), and verification.

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/random"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/i18n"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/sms"
	"github.com/bingo-project/bingo/internal/pkg/task"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// CodeScene 验证码场景
type CodeScene string

const (
	CodeSceneRegister      CodeScene = "register"
	CodeSceneResetPassword CodeScene = "reset_password"
	CodeSceneBind          CodeScene = "bind"
)

// CodeBiz 验证码业务接口
type CodeBiz interface {
	Send(ctx context.Context, account string, scene CodeScene) error
	Verify(ctx context.Context, account string, scene CodeScene, code string) error
}

type codeBiz struct {
	codeLength int
	codeTTL    int // 分钟
	codeWait   int // 分钟
}

func NewCodeBiz() CodeBiz {
	return &codeBiz{
		codeLength: 6,
		codeTTL:    5,
		codeWait:   1,
	}
}

func (b *codeBiz) Send(ctx context.Context, account string, scene CodeScene) error {
	accountType, err := DetectAccountType(account)
	if err != nil {
		return err
	}

	// 检查发送频率
	waitKey := fmt.Sprintf("verify_code_waiting:%s:%s", scene, account)
	if facade.Cache.Has(waitKey) {
		return errno.ErrTooManyRequests
	}

	// 生成验证码
	code := random.RandNumeral(b.codeLength)

	// 存储验证码
	codeKey := fmt.Sprintf("verify_code:%s:%s", scene, account)
	facade.Cache.Set(codeKey, code, time.Minute*time.Duration(b.codeTTL))

	// 设置发送间隔
	facade.Cache.Set(waitKey, "1", time.Minute*time.Duration(b.codeWait))

	// 发送验证码
	switch accountType {
	case AccountTypeEmail:
		return b.sendEmail(ctx, account, code, scene)
	case AccountTypePhone:
		return b.sendSMS(ctx, account, code, scene)
	}

	return nil
}

func (b *codeBiz) Verify(ctx context.Context, account string, scene CodeScene, code string) error {
	codeKey := fmt.Sprintf("verify_code:%s:%s", scene, account)
	stored := facade.Cache.Get(codeKey)
	if stored == nil || stored != code {
		return errno.ErrInvalidCode
	}

	// 验证成功后删除
	facade.Cache.Forget(codeKey)
	return nil
}

func (b *codeBiz) sendEmail(ctx context.Context, email, code string, scene CodeScene) error {
	lang := contextx.Lang(ctx)
	data := map[string]interface{}{
		"Code": code,
		"TTL":  b.codeTTL,
	}

	subject := i18n.T(lang, "code_"+string(scene)+"_subject", nil)
	msg := i18n.T(lang, "code_"+string(scene)+"_body", data)

	// Email task payload
	payload := &task.EmailVerificationCodePayload{
		To:      email,
		Subject: subject,
		Content: msg,
	}

	// Enqueue email task
	_, err := task.T.Queue(ctx, task.EmailVerificationCode, payload).Dispatch()
	if err != nil {
		log.C(ctx).Errorw("enqueue email task failed", "err", err)
		return err
	}

	log.C(ctx).Infow("sendEmail succeed", "email", email, "scene", scene, "lang", lang)
	return nil
}

func (b *codeBiz) sendSMS(ctx context.Context, phone, code string, scene CodeScene) error {
	if !sms.IsConfigured() {
		return errno.ErrSMSNotConfigured
	}

	lang := contextx.Lang(ctx)
	data := map[string]interface{}{
		"Code": code,
		"TTL":  b.codeTTL,
	}
	msg := i18n.T(lang, "code_"+string(scene)+"_body", data)

	// TODO: 实际发送短信，使用 msg 作为内容
	_ = msg
	log.C(ctx).Infow("sendSMS succeed", "phone", phone, "scene", scene, "lang", lang)
	return nil
}
